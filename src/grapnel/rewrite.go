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

package grapnel

import (
  "fmt"
  "regexp"
  "text/template"
  "bytes"
  toml "github.com/pelletier/go-toml"
  //log "grapnel/log"
)

type MatchMap map[string]*regexp.Regexp
type ReplaceMap map[string]*template.Template
type StringMap map[string]string

type RewriteRule struct {
  Matches MatchMap
  Replacements ReplaceMap
}

type RewriteRuleArray []*RewriteRule

func NewRewriteRule() *RewriteRule {
  return &RewriteRule {
    Matches: MatchMap{},
    Replacements: ReplaceMap{},
  }
}

func RewriteTemplate(tmpl string) (*template.Template, error) {
  return template.New("").Funcs(replaceFuncs).Parse(tmpl)
}

func TypeResolverRule(matchField, matchExpr, typeValue string) *RewriteRule {
  rule := NewRewriteRule()
  rule.Matches["type"] = regexp.MustCompile(`^$`)
  rule.Matches[matchField] = regexp.MustCompile(matchExpr)
  rule.Replacements["type"] = template.Must(RewriteTemplate(typeValue))
  return rule
}

func BuildRewriteRule(matches StringMap, replacements StringMap) *RewriteRule {
  rule := NewRewriteRule()
  for key, value := range matches {
    rule.Matches[key] = regexp.MustCompile(value)
  }
  for key, value := range replacements {
    rule.Replacements[key] = template.Must(RewriteTemplate(value))
  }
  return rule
}

func (self *RewriteRule) AddMatch(field, expr string) error {
  regex, err := regexp.Compile(expr)
  if err != nil {
    return err
  }
  self.Matches[field] = regex
  return nil
}

func (self *RewriteRule) AddReplacement(field, expr string) error {
  tmpl, err := template.New(field).Parse(expr)
  if err != nil {
    return err
  }
  self.Replacements[field] = tmpl
  return nil
}

// apply a match rule
func (self *RewriteRule) Apply(dep *Dependency) error {
  // match *all* expressions against the dependency
  depValues := dep.Flatten()
  for field, match := range self.Matches {
    if !match.MatchString(depValues[field]) {
      return nil // no match
    }
  }

  // generate new value map
  newValues := map[string]string{}
  writer := &bytes.Buffer{}
  for field, tmpl := range self.Replacements {
    writer.Reset()
    if err := tmpl.Execute(writer, depValues); err != nil {
      // TODO: need waaaay more context for this to be useful
      return fmt.Errorf("Error executing replacement rule: %v", err)
    }
    newValues[field] = writer.String()
  }

  // set up the new dependency
  if err := dep.SetValues(newValues); err != nil {
    return err
  }

  // return new dependency
  return nil
}

func (self RewriteRuleArray) Apply(dep *Dependency) error {
  for _, rule := range self {
    if err := rule.Apply(dep); err != nil {
      return err
    }
  }
  return nil
}

// Loads rewrite rules in a TOML file, specified by the filename argument.
// Returns an array of RewriteRules, or error.
func LoadRewriteRules(filename string) (RewriteRuleArray, error) {
  // load the config file
  tree, err := toml.LoadFile(filename)
  if err != nil {
    return nil, fmt.Errorf("%s: %s", filename, err)
  }

  // curry the filename and position into an error format function
  pos := toml.Position{}
  errorf := func(format string, values... interface{}) (RewriteRuleArray, error) {
    curriedFormat := filename + " " + pos.String() + ": " + format
    return nil, fmt.Errorf(curriedFormat, values...)
  }

  rules := RewriteRuleArray{}

  if rewriteTree := tree.Get("rewrite"); rewriteTree != nil {
    for _, ruleTree := range rewriteTree.([]*toml.TomlTree) {
      rule := NewRewriteRule()
      matchTree, ok := ruleTree.Get("match").(*toml.TomlTree)
      if !ok {
        pos = ruleTree.GetPosition("")
        return errorf("Expected 'match' subtree for rewrite rule")
      }
      replaceTree, ok := ruleTree.Get("replace").(*toml.TomlTree)
      if !ok {
        pos = ruleTree.GetPosition("")
        return errorf("Expected 'replace' subtree for rewrite rule")
      }
      for _, key := range matchTree.Keys() {
        matchString, ok := matchTree.Get(key).(string)
        if !ok {
          pos = matchTree.GetPosition(key)
          return errorf("Match expression must be a string value")
        }
        matchRegex, err := regexp.Compile(matchString)
        if err != nil {
          pos = matchTree.GetPosition(key)
          return errorf("Error compiling match expression: %s", err)
        }
        rule.Matches[key] = matchRegex
      }
      for _, key := range replaceTree.Keys() {
        replaceString, ok := replaceTree.Get(key).(string)
        if !ok {
          pos = replaceTree.GetPosition(key)
          return errorf("Replace expression must be a string value")
        }
        replaceTempl, err := RewriteTemplate(replaceString)
        if err != nil {
          pos = replaceTree.GetPosition(key)
          return errorf("Error compiling replace expression: %s", err)
        }
        rule.Replacements[key] = replaceTempl
      }
      rules = append(rules, rule)
    }
  }
  return rules, nil
}

func replace_Replace(value, expr, repl string) (string, error) {
  regex, err := regexp.Compile(expr)
  if err != nil {
    return "", err
  }
  return regex.ReplaceAllString(value, repl), nil
}

var replaceFuncs = template.FuncMap {
  "replace": replace_Replace,
}

var BasicRewriteRules = RewriteRuleArray {
  // generic rewrite for missing url
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
}
