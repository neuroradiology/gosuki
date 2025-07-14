//
//  Copyright (c) 2024-2025 Chakib Ben Ziane <contact@blob42.xyz>  and [`gosuki` contributors](https://github.com/blob42/gosuki/graphs/contributors).
//  All rights reserved.
//
//  SPDX-License-Identifier: AGPL-3.0-or-later
//
//  This file is part of GoSuki.
//
//  GoSuki is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  GoSuki is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with gosuki.  If not, see <http://www.gnu.org/licenses/>.
//

package main

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/blob42/gosuki/internal/utils"
	"github.com/blob42/gosuki/internal/webui"
	"github.com/blob42/gosuki/pkg/build"
	"github.com/blob42/gosuki/pkg/events"
	"github.com/blob42/gosuki/pkg/logging"
	"github.com/blob42/gosuki/pkg/manager"
	"github.com/blob42/gosuki/pkg/modules"
	"github.com/blob42/gosuki/pkg/profiles"
	"github.com/blob42/gosuki/pkg/watch"
)

const (
	maxWidth   = 80
	padding    = 2
	tickRate   = 20
	nLogLines  = 8
	helpHeight = 3
	statusChar = "●"
)

type module struct {
	modules.Module
	id modules.ModID

	progress progress.Model
	modState
}

type modState struct {
	totalCount uint
	curCount   uint
	enabled    bool
}

// Meta struct that holds progress for all profiles belonging to
// a browser. A browser flavour is considered its own browser instance
type browser struct {
	id string

	// Covers both profiles and flavours
	// profiles  []progress.Model
	instances     []modules.BrowserModule
	profileStates map[modules.Module]*modState
	progress      progress.Model
	modState
}

func updateBrowserProgress(b *browser, msg events.ProgressUpdateMsg) tea.Cmd {
	//TODO: handle update for profiled browser, or single instance
	profState, exists := b.profileStates[msg.Instance]
	if !exists {
		panic("instance does exist")
	}

	if msg.NewBk {
		profState.curCount += 1
		profState.totalCount = profState.curCount
		b.totalCount += 1
	} else {
		profState.curCount = msg.CurrentCount
		profState.totalCount = msg.Total
	}

	var newCurCount uint
	for _, prf := range b.profileStates {
		newCurCount += prf.curCount
	}
	b.curCount = newCurCount
	b.totalCount = 0
	for _, p := range b.profileStates {
		b.totalCount += p.totalCount
	}

	// logging.FDebugf("/tmp/gosuki_cur_count", "browser count: %d", newCurCount)

	percent := float64(newCurCount) / float64(b.totalCount)

	return b.progress.SetPercent(percent)

}

type winSize struct {
	width  int
	height int
}

type daemonState int

const (
	DaemonLoading = iota
	DaemonReady
	DaemonFail
)

type tuiModel struct {
	initFunc    initFunc
	logBuffer   *logging.TailBuffer
	manager     *manager.Manager
	modules     map[string]*module
	modKeys     []string
	browsers    map[string]*browser
	browserKeys []string
	windowSize  winSize
	keymap      keymap
	help        help.Model
	daemon      daemonState
}

type keymap struct {
	quit key.Binding
}

type ErrMsg error
type TickMsg time.Time
type DaemonStartedMsg struct{}
type DaemonStoppedMsg struct{}
type initFunc func(tea.Model) tea.Cmd

var (
	progressOpts = []progress.Option{
		progress.WithDefaultGradient(),
		progress.WithFillCharacters('━', '╺'),
		progress.WithoutPercentage(),
	}

	defaultTextColor = lipgloss.NewStyle().Foreground(
		lipgloss.AdaptiveColor{Light: "240", Dark: "255"},
	)

	logSectionStyle = lipgloss.NewStyle().
		// Background(lipgloss.Color("22")).
		MarginTop(2).
		Padding(1, 0, 0, 2).
		Width(maxWidth)

	titleStyle = defaultTextColor.
			PaddingLeft(2).
			MarginTop(1).
			MarginBottom(1)

	infoLabelStyle = defaultTextColor.
			AlignHorizontal(lipgloss.Right).
		// Background(lipgloss.Color("73")).
		MarginTop(1).
		Width(12).
		MarginLeft(2).
		MarginRight(3)

	labelStyle = defaultTextColor.
			MarginLeft(1).
			Width(10).
		// Background(lipgloss.Color("63")).
		Align(lipgloss.Right)

	urlCountStyle = defaultTextColor.
			PaddingLeft(1).
		// Background(lipgloss.Color("22")).
		Width(20)

	totalLabelStyle = defaultTextColor.
			Width(0).
			MarginLeft(labelStyle.GetWidth() + 2).
		// Background(lipgloss.Color("22")).
		MarginTop(1)

	ProgressSectionStyle = lipgloss.NewStyle().
				MarginTop(2)
		// Background(lipgloss.Color("24"))
		// Width(60)
		// MarginBottom(1)

	ProgressBarStyle = lipgloss.NewStyle().
				PaddingLeft(2)

	ProfilePathStyle = lipgloss.NewStyle().
				MarginLeft(2).
				Foreground(lipgloss.Color("245"))

	moduleStatusStyle = lipgloss.NewStyle().
				Bold(true)

	moduleOnStyle = moduleStatusStyle.
			PaddingRight(1).
			Foreground(lipgloss.Color("120"))

	moduleOffStyle = moduleStatusStyle.
			Foreground(lipgloss.Color("203"))

	helpStyle = lipgloss.NewStyle().Align(lipgloss.Bottom).
			PaddingLeft(4)
)

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*tickRate, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// Init implements tea.Model.
func (m tuiModel) Init() tea.Cmd {
	return tea.Batch(tea.ClearScreen, m.initFunc(m), tickCmd())
}

func (m tuiModel) HelpView() string {
	return "\n" + m.help.ShortHelpView([]key.Binding{
		m.keymap.quit,
	})
}

func setupModProgress(m tuiModel, r watch.WatchRunner) (tea.Model, tea.Cmd) {
	mod, ok := r.(modules.Module)
	if !ok {
		return m, nil
	}

	id := string(mod.ModInfo().ID)
	br, isB := mod.(modules.BrowserModule)
	if isB {
		// _, isProfMgr := br.(profiles.ProfileManager)

		if _, exists := m.browsers[id]; !exists {
			panic("missing browser map entry")
		}

		//TODO: custom profile handling
		// if isProfMgr {
		// TODO: show flavours
		// flavs := pm.ListFlavours()
		// for _, f := range flavs {
		// 	_, err := pm.GetProfiles(f.Name)
		// 	if err != nil {
		// 		panic(err)
		// 	}
		// }
		// }

		// track this instance (ie profile)
		m.browsers[id].instances = append(m.browsers[id].instances, br)

		// init profile state
		m.browsers[id].profileStates[br] = &modState{}
	} else {
		if _, exists := m.modules[id]; !exists {
			panic("missing module map entry")
		}

		m.modules[id] = &module{
			Module:   mod,
			progress: progress.New(progressOpts...),
			id:       mod.ModInfo().ID,
		}
	}

	return m, nil
}

// Update implements tea.Model.
func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			return m, tea.Sequence(func() tea.Msg {
				log.Info("stopping GoSuki ...")
				go m.manager.Shutdown()
				<-m.manager.Quit
				utils.CleanFiles()
				return nil
			}, tea.Quit)
		}

	case TickMsg:
		cmds := []tea.Cmd{tickCmd()}

		// watch for events
		cmds = append(cmds, func() tea.Msg { return <-events.TUIBus })

		// // update counters
		// for _, b := range m.browsers {
		// 	newTotal := 0
		// 	newCur := 0
		// 	for _, br := range b.instances {
		// 		if counter, ok := br.(parsing.Counter); ok {
		// 			newTotal += int(counter.Total())
		// 			newCur += int(counter.URLCount())
		// 		}
		// 	}
		// 	// Only update progress if there's a change
		// 	if newTotal != b.totalCount || newCur != b.curCount {
		// 		cmds = append(cmds, updateBrowserProgress(b, newCur, newTotal))
		// 	}
		// }

		return m, tea.Batch(cmds...)

	// New module instance
	case events.RunnerStarted:
		return setupModProgress(m, msg.WatchRunner)

	case events.StartedLoadingMsg:
		_, isBr := m.browsers[string(msg.ID)]

		// simple module
		if !isBr {
			_, isMod := m.modules[string(msg.ID)]
			if !isMod {
				errMsg := fmt.Sprintf("%s: module not recognized", msg.ID)
				panic(errMsg)
			}
			m.modules[string(msg.ID)].enabled = true
		} else {
			m.browsers[string(msg.ID)].enabled = true
		}

		return m, nil

	case events.ProgressUpdateMsg:

		// logging.FDebugf("/tmp/gosuki_progress", "%#v", msg)
		// browser
		if b, ok := m.browsers[string(msg.ID)]; ok {
			return m, updateBrowserProgress(b, msg)

			// simple module
		} else if mod, ok := m.modules[string(msg.ID)]; ok {
			m.modules[string(msg.ID)].enabled = true
			mod.curCount = msg.CurrentCount
			mod.totalCount = msg.Total
			percent := float64(mod.curCount) / float64(mod.totalCount)
			return m, mod.progress.SetPercent(percent)
		} else {
			log.Error("unrecognized module", "module", msg.ID)
		}

	case tea.WindowSizeMsg:
		titleStyle = titleStyle.Width(msg.Width)

		// TODO: responsive
		for _, m := range m.browsers {
			m.progress.Width = min(int(math.Min(float64(msg.Width/2), 80)), maxWidth)
		}

		for _, m := range m.modules {
			m.progress.Width = min(int(math.Min(float64(msg.Width/2), 80)), maxWidth)
		}

		m.windowSize.height = msg.Height
		m.windowSize.width = msg.Width

		logSectionStyle = logSectionStyle.Width(msg.Width)
		// ProgressSectionStyle = ProgressBarStyle.Width(msg.Width - 10)
		return m, nil

	case ErrMsg:
		fmt.Printf("tui error: %s", msg.Error())
		return m, tea.Quit

	case DaemonStartedMsg:
		m.daemon = DaemonReady

	case progress.FrameMsg:
		cmds := []tea.Cmd{}
		for _, b := range m.browsers {
			progressModel, cmd := b.progress.Update(msg)
			b.progress = progressModel.(progress.Model)
			cmds = append(cmds, cmd)
		}

		for _, m := range m.modules {
			progressModel, cmd := m.progress.Update(msg)
			m.progress = progressModel.(progress.Model)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	default:
		return m, nil
	}

	return m, nil
}

// View implements tea.Model.
func (m tuiModel) View() string {
	doc := strings.Builder{}

	doc.WriteString(titleStyle.Render(fmt.Sprintf("gosuki %s", build.Version())))

	// get web UI status
	var webUIOK bool
	for name := range m.manager.Units() {
		if strings.HasPrefix(name, "webui") {
			webUIOK = true
		}
	}

	var webUILabel lipgloss.Style
	if webUIOK {
		webUILabel = moduleOnStyle
	} else {
		webUILabel = moduleOffStyle
	}

	uiSection := strings.Builder{}
	uiSection.WriteString(infoLabelStyle.Render(
		webUILabel.Render(statusChar),
		defaultTextColor.Render("web ui:"),
	))
	// uiSection.WriteString(infoLabelStyle.Render(
	// 	fmt.Sprintf("%s web ui :", webUILabel)))
	uiSection.WriteString(defaultTextColor.Render(fmt.Sprintf("http://%s", webui.BindAddr)))

	progressSection := strings.Builder{}

	var totalURLCount uint

	// Browser modules
	for _, name := range m.browserKeys {
		br, ok := m.browsers[name]
		if !ok || !br.enabled {
			continue
		}
		totalURLCount += uint(br.totalCount)
		progressSection.WriteString(labelStyle.Render(name))
		progressSection.WriteString(ProgressBarStyle.Render(br.progress.View()))
		progressSection.WriteString(urlCountStyle.Render(
			fmt.Sprintf("%4d/%-4d", br.curCount, br.totalCount)),
		)

		// handle multiple browser instances such as profiles and flavours
		if len(br.instances) > 1 {
			for _, brr := range br.instances {
				bpm, ok := brr.(profiles.ProfileManager)
				if !ok {
					continue
				}
				profile := bpm.GetProfile()
				// if profile == nil {
				// 	continue
				// }

				progressSection.WriteString("\n")

				labelLineStyle := labelStyle.
					Foreground(lipgloss.Color("99"))

				progressSection.WriteString(labelLineStyle.Render("└─") + " ")
				progressSection.WriteString(defaultTextColor.Render(profile.Name))
				progressSection.WriteString(ProfilePathStyle.Render(profile.BaseDir))

				// if cfg.Flavour != nil {
				// 	progressSection.WriteString(cfg.Flavour.Name + " ")
				// }
			}
			progressSection.WriteByte('\n')
		} else {
			progressSection.WriteByte('\n')
		}
	}

	// Simple modules
	for _, name := range m.modKeys {
		mod := m.modules[name]
		if !mod.enabled {
			continue
		}
		totalURLCount += uint(mod.totalCount)
		progressSection.WriteString(labelStyle.Render(name))
		progressSection.WriteString(ProgressBarStyle.
			Render(mod.progress.View()))
		progressSection.WriteString(urlCountStyle.Render(
			fmt.Sprintf("%4d/%-4d", mod.curCount, mod.totalCount)),
		)
		progressSection.WriteByte('\n')
	}

	//WIP:
	doc.WriteString(uiSection.String())
	doc.WriteString(infoLabelStyle.Render("modules:"))
	doc.WriteString(defaultTextColor.Render(fmt.Sprintf("%d", len(m.modules)+len(m.browsers))))
	doc.WriteString(ProgressSectionStyle.
		// Height(m.windowSize.height / 2).
		Render(progressSection.String()))

	totalLabelStyle = totalLabelStyle.MarginLeft(labelStyle.GetWidth())
	doc.WriteString(totalLabelStyle.Render(fmt.Sprintf("%d bookmarks loaded", totalURLCount)))
	// doc.WriteString(fmt.Sprintf("%d", totalUrlCount))

	doc.WriteString(logSectionStyle.Render(strings.Join(m.logBuffer.Lines(), "\n")) + "\n")

	doc.WriteString(helpStyle.Render(m.HelpView()))
	return doc.String()
}

type tui struct {
	model tuiModel
	opts  []tea.ProgramOption
}

func NewTUI(
	initFunc initFunc,
	manager *manager.Manager,
	opts ...tea.ProgramOption,
) *tui {

	mods := map[string]*module{}
	browsers := make(map[string]*browser)
	browserKeys := []string{}
	modKyes := []string{}

	for _, mod := range modules.GetModules() {
		modInfo := mod.ModInfo()
		name := string(modInfo.ID)

		// Extract module/browser name, convert flavours to name as well
		b, isBrowser := modInfo.New().(modules.BrowserModule)
		if isBrowser {
			browserKeys = append(browserKeys, b.Config().Name)
			browsers[b.Config().Name] = &browser{
				id:       b.Config().Name,
				progress: progress.New(progressOpts...),
				// profiles:  []progress.Model{},
				instances:     []modules.BrowserModule{},
				profileStates: map[modules.Module]*modState{},
			}
		} else {
			modKyes = append(modKyes, name)
			mods[name] = &module{
				id:       modules.ModID(name),
				Module:   mod,
				progress: progress.New(progressOpts...),
			}
		}

		modLabelWidth := labelStyle.GetWidth()
		modLabelWidth = int(math.Max(float64(modLabelWidth), float64(len(name))))
		labelStyle = labelStyle.Width(modLabelWidth)
	}

	return &tui{
		model: tuiModel{
			initFunc:    initFunc,
			manager:     manager,
			logBuffer:   logging.NewTailBuffer(nLogLines),
			modules:     mods,
			browsers:    browsers,
			browserKeys: browserKeys,
			modKeys:     modKyes,
			windowSize:  winSize{},
			keymap: keymap{
				quit: key.NewBinding(
					key.WithKeys("q", "esc", "ctrl+c"),
					key.WithHelp("q/esc", "quit"),
				),
			},
			help:   help.New(),
			daemon: DaemonLoading,
		},
		opts: opts,
	}
}

func (tui *tui) Run() error {
	_, err := tea.NewProgram(tui.model, tui.opts...).Run()
	if err != nil {
		return errors.New("could not start TUI")
	}
	return nil
}
