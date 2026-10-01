// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/google/ax/pkg/apis/v1alpha1"
	"gopkg.in/yaml.v3"
)

// repoRoot returns the directory holding go.mod, so the tests below can reach
// docs/ and examples/ regardless of where `go test` is invoked from.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find go.mod above the test working directory")
		}
		dir = parent
	}
}

// mdLink matches the target of an inline Markdown link, capturing only the
// path portion. Anchors and blank targets are left out because neither can be
// resolved by stat.
var mdLink = regexp.MustCompile(`\]\(([^)\s]+)\)`)

// TestDocsLinksResolve checks that every relative link in the documentation
// points at a file that exists. A link to a nonexistent example is the kind of
// drift that only surfaces when a reader follows it.
func TestDocsLinksResolve(t *testing.T) {
	root := repoRoot(t)
	files, err := filepath.Glob(filepath.Join(root, "docs", "*.md"))
	if err != nil {
		t.Fatalf("glob docs: %v", err)
	}
	files = append(files, filepath.Join(root, "README.md"))

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("read %s: %v", file, err)
			}
			for _, m := range mdLink.FindAllStringSubmatch(string(data), -1) {
				target := m[1]
				if i := strings.IndexAny(target, "#"); i >= 0 {
					target = target[:i]
				}
				if target == "" {
					continue
				}
				if scheme := strings.Index(target, "://"); scheme >= 0 || strings.HasPrefix(target, "mailto:") {
					continue // remote, nothing to resolve on disk
				}
				if filepath.IsAbs(target) {
					continue
				}
				path := filepath.Join(filepath.Dir(file), target)
				if _, err := os.Stat(path); err != nil {
					t.Errorf("%s links to %s, which does not exist", file, target)
				}
			}
		})
	}
}

// TestExamplesAreValidManifests checks that every file in examples/ decodes
// into the API types and passes the same validation the server applies, so an
// example cannot drift away from the schema it documents.
func TestExamplesAreValidManifests(t *testing.T) {
	root := repoRoot(t)
	paths, err := filepath.Glob(filepath.Join(root, "examples", "*.yaml"))
	if err != nil {
		t.Fatalf("glob examples: %v", err)
	}
	if len(paths) == 0 {
		t.Fatal("no example manifests found")
	}

	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}

			decoder := yaml.NewDecoder(strings.NewReader(string(data)))
			docs := 0
			for {
				var doc yaml.Node
				if err := decoder.Decode(&doc); err != nil {
					if err.Error() == "EOF" {
						break
					}
					t.Fatalf("decoding document %d: %v", docs+1, err)
				}
				if len(doc.Content) == 0 {
					continue // empty document, e.g. a trailing "---"
				}
				docs++

				var head struct {
					Kind string `yaml:"kind"`
				}
				if err := doc.Content[0].Decode(&head); err != nil {
					t.Fatalf("document %d: reading kind: %v", docs, err)
				}

				switch head.Kind {
				case v1alpha1.KindTask:
					var task v1alpha1.Task
					if err := doc.Content[0].Decode(&task); err != nil {
						t.Fatalf("document %d: decoding Task: %v", docs, err)
					}
					if err := v1alpha1.ValidateTask(&task); err != nil {
						t.Errorf("document %d: ValidateTask: %v", docs, err)
					}
				case v1alpha1.KindWorkspace:
					var ws v1alpha1.Workspace
					if err := doc.Content[0].Decode(&ws); err != nil {
						t.Fatalf("document %d: decoding Workspace: %v", docs, err)
					}
					if err := v1alpha1.ValidateWorkspace(&ws); err != nil {
						t.Errorf("document %d: ValidateWorkspace: %v", docs, err)
					}
				case v1alpha1.KindModel:
					var model v1alpha1.Model
					if err := doc.Content[0].Decode(&model); err != nil {
						t.Fatalf("document %d: decoding Model: %v", docs, err)
					}
					if err := v1alpha1.ValidateModel(&model); err != nil {
						t.Errorf("document %d: ValidateModel: %v", docs, err)
					}
				default:
					t.Errorf("document %d: unsupported kind %q", docs, head.Kind)
				}
			}
			if docs == 0 {
				t.Error("example contains no documents")
			}
		})
	}
}
