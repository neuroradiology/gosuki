//
//  Copyright (c) 2024 Chakib Ben Ziane <contact@blob42.xyz>  and [`gosuki` contributors](https://github.com/blob42/gosuki/graphs/contributors).
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

// Package marktab handles reading and parsing of marktab files, inspired by the crontab file format. Marktab files are text files that define rules and actions to execute when gosuki detects a bookmark with tags matching the defined rule.
//
// # Marktab Format:
//
// Each line in a marktab file represents a rule.
// A rule consists of three fields: trigger, pattern, and command.
// The syntax for each line is as follows:
//
//	#  *  *   *
//	#  |  |   |_____ shell command to execute
//	#  |  |_________ pattern to match on the url or title
//	#  |____________ trigger keyword to detect on the tags
//
// Example:
//
//	notify		.*		my_notify_script
//
// The syntax of each line expects an expression made of three field: trigger, pattern and command.
//
// # Trigger:
//
// The keyword to detect in the bookmark tags. When gosuki parses bookmark tags and finds this trigger keyword, it evaluates the pattern rule. If the pattern matches, the corresponding command is executed.
//
// # Pattern:
//
// A regular expression used for matching against a part of the bookmark URL or title. Once a trigger is detected, the pattern is evaluated to determine if there's a match with the bookmark data.
//
// # Command:
//
// The shell command to execute when both the trigger and pattern are matched in a bookmark tag. This command can be any valid shell command and it allows for flexibility in performing various actions.
package marktab

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/blob42/gosuki/internal/utils"
	"github.com/blob42/gosuki/pkg/logging"
)

// MarkTab represents a collection of rules defined in the marktab file as lines.
type MarkTab struct {
	Rules []Rule // Rules contains all the parsed rules from the marktab file.
}

type Rule struct {
	Trigger string // keyword to detect in the bookmark tags
	Pattern string // regular expression used for matching against the bookmark URL or title.
	Command string // shell command to execute when both the trigger and pattern match the bookmark tags.

	empty bool // empty is an unexported field indicating whether the rule is empty.
}

const marktabPath = "~/.config/gosuki/marktab"

type ErrorType int

// marktab parsing errors
const (
	// No error
	OK ErrorType = iota

	ErrBadPattern

	ErrBadTrigger

	ErrBadRule
)

const InvalidFormat = "invalid format"

var (
	log         = logging.GetLogger("marktab")
	CachedRules *MarkTab
)

type MarktabError struct {
	ErrorType
	Rule    *Rule
	Context string
	err     error
}

func errInvalidFormat(context string, a ...any) error {
	if len(a) > 0 {
		return fmt.Errorf("\n %s\n\ninvalid format: %s", context, a[0])
	}

	return fmt.Errorf("\n %s\n\ninvalid format", context)
}

func errBadPattern(pat string, context string, err error) error {
	return MarktabError{
		ErrorType: ErrBadPattern,
		Rule:      &Rule{Pattern: pat},
		Context:   context,
		err:       err,
	}.Error()
}

// TODO:
func (mte MarktabError) Error() error {
	var outErr string
	switch mte.ErrorType {
	case ErrBadPattern:
		outErr = fmt.Sprintf("\n %s\n\ninvalid pattern: `%s'", mte.Context, mte.Rule.Pattern)
	case ErrBadTrigger:
		outErr = fmt.Sprintf(" %s\ninvalid trigger: `%s'", mte.Context, mte.Rule.Trigger)
	case ErrBadRule:
		outErr = fmt.Sprintf(" %s\ninvalid rule", mte.Context)
	}
	if mte.err != nil {
		return fmt.Errorf("%s : %w", outErr, mte.err)
	}

	return fmt.Errorf("%s", outErr)
}

func PreloadRules() error {
	if CachedRules == nil {
		CachedRules = &MarkTab{}
		return CachedRules.LoadMarktabs()
	}
	return nil
}

func (mt *MarkTab) LoadMarktabs() error {
	path, err := utils.ExpandPath(marktabPath)
	if os.IsNotExist(err) {
		log.Warnf("skipping marktab, not found: %v", marktabPath)
		return nil

	} else if err != nil {
		return fmt.Errorf("reading %s : %w ", marktabPath, err)
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		rule, err := parseLine(line)
		if err != nil {
			return err
		}
		if !rule.empty {
			mt.Rules = append(mt.Rules, rule)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func parseLine(line string) (Rule, error) {
	line = skipComments(line)
	if len(line) == 0 {
		return Rule{empty: true}, nil
	}

	fields := strings.Fields(line)

	// if len(fields) != 3 {
	// 	return Rule{}, errInvalidFormat(line, "wrong number of fields")
	// }

	trigger := fields[0]
	pattern := fields[1]
	command := strings.Join(fields[2:], " ")

	// Validate pattern (basic check, can be extended based on requirements)
	_, err := regexp.Compile(pattern)
	if err != nil {
		return Rule{}, errBadPattern(pattern, line, err)
	}

	return Rule{Trigger: trigger, Pattern: pattern, Command: command}, nil
}

func skipComments(line string) string {
	if idx := strings.Index(line, "#"); idx != -1 {
		return line[:idx]
	}
	return line
}
