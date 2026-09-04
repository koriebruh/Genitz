package generator

import (
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// validLicenseKinds are the values licenseContent recognizes, plus "" (no
// license) — used by ValidLicenseKind so callers can reject a typo'd
// --license flag instead of silently generating nothing.
var validLicenseKinds = map[string]bool{"": true, "mit": true, "apache-2.0": true}

// ValidLicenseKind reports whether kind is a value licenseContent
// recognizes ("" for none, "mit", or "apache-2.0") — main.go uses this to
// fail fast on an unrecognized --license value instead of silently
// generating no LICENSE file.
func ValidLicenseKind(kind string) bool {
	return validLicenseKinds[kind]
}

// GitConfigUserName returns `git config --get user.name`'s value, or "" if
// git isn't on PATH, isn't configured, or the lookup otherwise fails —
// callers treat "" as "no default available" and fall back to a
// placeholder, so a failure here is never fatal.
func GitConfigUserName() string {
	out, err := exec.Command("git", "config", "--get", "user.name").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// ResolveLicenseHolder returns the best available default copyright holder:
// git config user.name first, then the persisted genitz config's author
// (`genitz config set author "..."`), else "" — licenseContent falls back
// to the [COPYRIGHT HOLDER] placeholder from there.
func ResolveLicenseHolder() string {
	if name := GitConfigUserName(); name != "" {
		return name
	}
	cfg, _ := LoadConfig()
	return cfg.Author
}

// licenseContent returns the LICENSE body for kind ("mit" or "apache-2.0"),
// and false for "" / "none" / anything unrecognized — no license file is
// generated in that case rather than guessing. holder fills the copyright
// line when non-empty (typically GitConfigUserName's result); an empty
// holder leaves the [COPYRIGHT HOLDER] bracket placeholder for the user to
// fill in by hand, since there's no reliable source for a real name when
// git isn't configured. The year is always the current one.
func licenseContent(kind, holder string) (string, bool) {
	var tmpl string
	switch kind {
	case "mit":
		tmpl = mitLicense
	case "apache-2.0":
		tmpl = apacheLicense
	default:
		return "", false
	}
	tmpl = strings.ReplaceAll(tmpl, "[year]", strconv.Itoa(time.Now().Year()))
	if holder != "" {
		tmpl = strings.ReplaceAll(tmpl, "[COPYRIGHT HOLDER]", holder)
	}
	return tmpl, true
}

const mitLicense = `MIT License

Copyright (c) [year] [COPYRIGHT HOLDER]

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
`

const apacheLicense = `                                 Apache License
                           Version 2.0, January 2004
                        https://www.apache.org/licenses/

   Copyright [year] [COPYRIGHT HOLDER]

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       https://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
`
