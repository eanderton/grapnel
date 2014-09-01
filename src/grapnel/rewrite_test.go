package grapnel
/*
Copyright (c) 2014 Eric Anderton <eric.t.anderton@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

import (
  "testing"
  "reflect"
  "regexp"
  "text/template"
  url "grapnel/url"
  log "grapnel/log"
)

// test standard rewrite rules
func TestRewrite(t *testing.T) {
  rules := RewriteRuleArray{}
  rules = append(rules, BasicRewriteRules...)
  rules = append(rules, GitRewriteRules...)

  for _,test := range []struct {
    Src *Dependency
    Dst *Dependency
  } {
    {
      Src: &Dependency{
        Import: "gopkg.in/foo/bar.v3",
      },
      Dst: &Dependency{
        Import: "gopkg.in/foo/bar.v3",
        Url: url.MustParse("http://github.com/foo/bar"),
        Branch: "v3",
        Type: "git",

      },
    },
  } {
    if err := rules.Apply(test.Src); err != nil {
      t.Errorf("Error during replacement %v; Src: %v", err, test.Src.Flatten())
    }
    if !test.Src.Equal(test.Dst) {
      t.Errorf("Error during replacement Src: %v; Dst: %v",
        test.Src.Flatten(), test.Dst.Flatten())
    }
  }
}

func TestLoadRewriteRules(t *testing.T) {
  log.SetGlobalLogLevel(log.DEBUG)

  testRules := []*RewriteRule{
    &RewriteRule{
      Matches: MatchMap{
        "import": regexp.MustCompile(`.+`),
        "url":    regexp.MustCompile(`^$`),
      },
      Replacements: ReplaceMap{
        "url":    template.Must(RewriteTemplate(`http://{{.import}}`)),
      },
    },

    // generic rewrite for missing import
    &RewriteRule{
      Matches: MatchMap{
        "import": regexp.MustCompile(`^$`),
        "url":    regexp.MustCompile(`.+`),
      },
      Replacements: ReplaceMap{
        "import":    template.Must(RewriteTemplate(`{{.host}}/{{.path}}`)),
      },
    },
    // rewrite rules for misc git resolvers
    SimpleRewriteRule("scheme", `git`,           "type", `git`),
    SimpleRewriteRule("path",   `.*\.git`,       "type", `git`),
    SimpleRewriteRule("import", `github.com/.*`, "type", `git`),
    SimpleRewriteRule("host",   `github.com`,    "type", `git`),

    // rewrite rules for gopkg.in
    &RewriteRule{
      Matches: MatchMap{
        "host": regexp.MustCompile(`gopkg\.in`),
      },
      Replacements: ReplaceMap{
        "branch": template.Must(RewriteTemplate(`{{replace .path "^.*\\.(.*)$" "$1"}}`)),
        "path":   template.Must(RewriteTemplate(`{{replace .path "^(.*)\\..*$" "$1"}}`)),
        "host":   template.Must(RewriteTemplate(`github.com`)),
        "type":   template.Must(RewriteTemplate(`git`)),
      },
    },
  }

  rules, err := LoadRewriteRules("testfiles/grapnelrc.toml")
  if err != nil {
    t.Errorf("Error loading rules")
    t.Errorf("%v", err)
    return
  }
  if len(rules) != len(testRules) {
    t.Errorf("Array length for rules do not match: %d vs %d", len(rules), len(testRules))
    return
  }

  for idx, testRule := range testRules {
    rule := rules[idx]
    if len(rule.Matches) != len(testRule.Matches) {
      t.Errorf("Match array length for test %d do not match", idx)
      t.Logf("testRule: %v", testRule.Matches)
      t.Logf("rule: %v", rule.Matches)
    }
    if len(rule.Replacements) != len(testRule.Replacements) {
      t.Errorf("Replacement array length for test %d do not match", idx)
      t.Logf("testRule: %v", testRule.Replacements)
      t.Logf("rule: %v", rule.Replacements)
    }
    for key, testMatch := range testRule.Matches {
      if match, ok := rule.Matches[key]; !ok {
        t.Errorf("Test %d - Rule match key %s is missing", idx, key)
      } else if !reflect.DeepEqual(match, testMatch) {
        t.Errorf("Test %d - Rule match for key %s does not equal expected value", idx, key)
      }
    }
/* TODO: figure out some way to verify the templates have parsed and match
    for key, testReplace := range testRule.Replacements {
      if replace, ok := rule.Replacements[key]; !ok {
        t.Errorf("Test %d - Rule replace key %s is missing", idx, key)
      } else if !reflect.DeepEqual(replace, testReplace) {
        t.Errorf("Test %d - Rule replace for key %s does not equal expected value", idx, key)
      }
    }
*/
  }
}
