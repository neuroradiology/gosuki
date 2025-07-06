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

package marktab

import (
	"regexp"

	"github.com/blob42/gosuki"
)

// Match checks if a bookmark matches the rule based on its title, URL, and
// tags. It returns true if both the trigger (part of the tag) or the pattern
// match is satisfied, otherwise it returns false.
func (rule Rule) Match(bk *gosuki.Bookmark) bool {
	var triggerMatch bool
	if bk == nil {
		return false
	}
	for _, tag := range bk.Tags {
		if rule.Trigger == tag {
			triggerMatch = true
		}
	}
	patternMatch := rule.Pattern == "" || regexp.MustCompile(rule.Pattern).MatchString(bk.URL) || regexp.MustCompile(rule.Pattern).MatchString(bk.Title)
	if triggerMatch && patternMatch {
		return true
	}
	return false
}
