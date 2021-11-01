/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package env

import (
	"context"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestEnv_New(t *testing.T) {
	e := newTestEnv()
	if e.ctx == nil {
		t.Error("missing default context")
	}

	if len(e.actions) != 0 {
		t.Error("unexpected actions found")
	}

	if e.cfg.Namespace() != "" {
		t.Error("unexpected envconfig.Namespace value")
	}
}

func TestEnv_APIMethods(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*testing.T) *testEnv
		roles map[actionRole]int
	}{
		{
			name: "empty actions",
			setup: func(t *testing.T) *testEnv {
				return newTestEnv()
			},
			roles: map[actionRole]int{roleSetup: 0, roleBeforeTest: 0, roleAfterTest: 0, roleFinish: 0},
		},
		{
			name: "setup actions",
			setup: func(t *testing.T) *testEnv {
				env := newTestEnv()
				env.Setup(func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
					return ctx, nil
				}).Setup(func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
					return ctx, nil
				})
				return env
			},
			roles: map[actionRole]int{roleSetup: 2, roleBeforeTest: 0, roleAfterTest: 0, roleFinish: 0},
		},
		{
			name: "before actions",
			setup: func(t *testing.T) *testEnv {
				env := newTestEnv()
				env.BeforeEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					return ctx, nil
				})
				return env
			},
			roles: map[actionRole]int{roleSetup: 0, roleBeforeTest: 1, roleAfterTest: 0, roleFinish: 0},
		},
		{
			name: "after actions",
			setup: func(t *testing.T) *testEnv {
				env := newTestEnv()
				env.AfterEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					return ctx, nil
				})
				return env
			},
			roles: map[actionRole]int{roleSetup: 0, roleBeforeTest: 0, roleAfterTest: 1, roleFinish: 0},
		},
		{
			name: "finish actions",
			setup: func(t *testing.T) *testEnv {
				env := newTestEnv()
				env.Finish(func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
					return ctx, nil
				})
				return env
			},
			roles: map[actionRole]int{roleSetup: 0, roleBeforeTest: 0, roleAfterTest: 0, roleFinish: 1},
		},
		{
			name: "all actions",
			setup: func(t *testing.T) *testEnv {
				env := newTestEnv()
				env.Setup(func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
					return ctx, nil
				}).BeforeEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					return ctx, nil
				}).AfterEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					return ctx, nil
				}).Finish(func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
					return ctx, nil
				})
				return env
			},
			roles: map[actionRole]int{roleSetup: 1, roleBeforeTest: 1, roleAfterTest: 1, roleFinish: 1},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env := test.setup(t)
			for role, count := range test.roles {
				actual := len(env.getActionsByRole(role))
				if actual != count {
					t.Errorf("unexpected number of actions %d for role %d", actual, role)
				}
			}
		})
	}
}

func TestEnv_Test(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		setup    func(*testing.T, context.Context) []string
		expected []string
	}{
		{
			name: "feature only",
			ctx:  context.TODO(),
			expected: []string{
				"test-feat",
			},
			setup: func(t *testing.T, ctx context.Context) (val []string) {
				env := newTestEnv()
				f := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat")
					return ctx
				})
				env.Test(t, f.Feature())
				return
			},
		},
		{
			name: "filtered feature",
			ctx:  context.TODO(),
			expected: []string{
				"test-feat",
			},
			setup: func(t *testing.T, ctx context.Context) (val []string) {
				env := NewWithConfig(envconf.New().WithFeatureRegex("test-feat"))
				f := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat")
					return ctx
				})
				env.Test(t, f.Feature())

				env2 := NewWithConfig(envconf.New().WithFeatureRegex("skip-me"))
				f2 := features.New("test-feat-2").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat-2")
					return ctx
				})
				env2.Test(t, f2.Feature())

				return
			},
		},
		{
			name: "with before-test",
			ctx:  context.TODO(),
			expected: []string{
				"before-each-test",
				"test-feat",
			},
			setup: func(t *testing.T, ctx context.Context) (val []string) {
				env := newTestEnv()
				env.BeforeEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					val = append(val, "before-each-test")
					return ctx, nil
				})
				f := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat")
					return ctx
				})
				env.Test(t, f.Feature())
				return
			},
		},
		{
			name: "with after-test",
			ctx:  context.TODO(),
			expected: []string{
				"before-each-test",
				"test-feat",
				"after-each-test",
			},
			setup: func(t *testing.T, ctx context.Context) (val []string) {
				env := newTestEnv()
				env.AfterEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					val = append(val, "after-each-test")
					return ctx, nil
				}).BeforeEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					val = append(val, "before-each-test")
					return ctx, nil
				})
				f := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat")
					return ctx
				})
				env.Test(t, f.Feature())
				return
			},
		},
		{
			name: "with before-after-test",
			ctx:  context.TODO(),
			expected: []string{
				"test-feat",
				"after-each-test",
			},
			setup: func(t *testing.T, ctx context.Context) (val []string) {
				env := newTestEnv()
				env.AfterEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					val = append(val, "after-each-test")
					return ctx, nil
				})
				f := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat")
					return ctx
				})
				env.Test(t, f.Feature())
				return
			},
		},
		{
			name: "filter assessment",
			ctx:  context.TODO(),
			expected: []string{
				"add-1",
				"add-2",
			},
			setup: func(t *testing.T, ctx context.Context) (val []string) {
				val = []string{}
				env := NewWithConfig(envconf.New().WithAssessmentRegex("add-*"))
				f := features.New("test-feat").
					Assess("add-one", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
						val = append(val, "add-1")
						return ctx
					}).
					Assess("add-two", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
						val = append(val, "add-2")
						return ctx
					}).
					Assess("take-one", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
						val = append(val, "take-1")
						return ctx
					})
				env.Test(t, f.Feature())
				return
			},
		},
		{
			name: "context value propagation with before, during, and after test",
			ctx:  context.TODO(),
			expected: []string{
				"before-each-test",
				"test-feat",
				"after-each-test",
			},
			setup: func(t *testing.T, ctx context.Context) []string {
				env, err := NewWithContext(context.WithValue(ctx, &ctxTestKeyString{}, []string{}), envconf.New())
				if err != nil {
					t.Fatal(err)
				}
				env.BeforeEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					// update before test
					val, ok := ctx.Value(&ctxTestKeyString{}).([]string)
					if !ok {
						t.Fatal("context value was not []string")
					}
					val = append(val, "before-each-test")
					return context.WithValue(ctx, &ctxTestKeyString{}, val), nil
				})
				env.AfterEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					// update after the test
					val, ok := ctx.Value(&ctxTestKeyString{}).([]string)
					if !ok {
						t.Fatal("context value was not []string")
					}
					val = append(val, "after-each-test")
					return context.WithValue(ctx, &ctxTestKeyString{}, val), nil
				})
				f := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val, ok := ctx.Value(&ctxTestKeyString{}).([]string)
					if !ok {
						t.Fatal("context value was not []string")
					}
					val = append(val, "test-feat")

					return context.WithValue(ctx, &ctxTestKeyString{}, val)
				})

				env.Test(t, f.Feature())
				return env.(*testEnv).ctx.Value(&ctxTestKeyString{}).([]string)
			},
		},
		{
			name:     "no features specified",
			ctx:      context.TODO(),
			expected: []string{},
			setup: func(t *testing.T, ctx context.Context) (val []string) {
				env := newTestEnv()
				env.Test(t)
				return
			},
		},
		{
			name: "multiple features",
			ctx:  context.TODO(),
			expected: []string{
				"test-feature-1",
				"test-feature-2",
			},
			setup: func(t *testing.T, ctx context.Context) (val []string) {
				env := newTestEnv()
				f1 := features.New("test-feat-1").
					Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
						val = append(val, "test-feature-1")
						return ctx
					})

				f2 := features.New("test-feat-2").
					Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
						val = append(val, "test-feature-2")
						return ctx
					})

				env.Test(t, f1.Feature(), f2.Feature())
				return
			},
		},
		{
			name: "multiple features with before-after-test",
			ctx:  context.TODO(),
			expected: []string{
				"before-each-test",
				"test-feat-1",
				"test-feat-2",
				"after-each-test",
			},
			setup: func(t *testing.T, ctx context.Context) (val []string) {
				env := newTestEnv()
				val = []string{}
				env.BeforeEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					val = append(val, "before-each-test")
					return ctx, nil
				})
				env.AfterEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					val = append(val, "after-each-test")
					return ctx, nil
				})
				f1 := features.New("test-feat-1").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat-1")
					return ctx
				})
				f2 := features.New("test-feat-2").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat-2")
					return ctx
				})
				env.Test(t, f1.Feature(), f2.Feature())
				return
			},
		},
		{
			name: "with before-and-after features",
			ctx:  context.TODO(),
			expected: []string{
				"before-each-feature",
				"test-feat-1",
				"after-each-feature",
				"before-each-feature",
				"test-feat-2",
				"after-each-feature",
			},
			setup: func(t *testing.T, ctx context.Context) []string {
				env := newTestEnv()
				val := []string{}
				env.BeforeEachFeature(func(ctx context.Context, _ *envconf.Config, info features.Feature) (context.Context, error) {
					val = append(val, "before-each-feature")
					return ctx, nil
				}).AfterEachFeature(func(ctx context.Context, _ *envconf.Config, info features.Feature) (context.Context, error) {
					val = append(val, "after-each-feature")
					return ctx, nil
				})
				f1 := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat-1")
					return ctx
				})
				f2 := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat-2")
					return ctx
				})
				env.Test(t, f1.Feature(), f2.Feature())
				return val
			},
		}, {
			name: "before-and-after features unable to mutate feature",
			ctx:  context.TODO(),
			expected: []string{
				"before-each-feature",
				"test-feat-1",
				"after-each-feature",
				"before-each-feature",
				"test-feat-2",
				"after-each-feature",
			},
			setup: func(t *testing.T, ctx context.Context) []string {
				env := newTestEnv()
				val := []string{}
				env.BeforeEachFeature(func(ctx context.Context, _ *envconf.Config, info features.Feature) (context.Context, error) {
					val = append(val, "before-each-feature")
					t.Logf("%#v, len(steps)=%v step[0].Name: %v\n", info, len(info.Steps()), info.Steps()[0].Name())

					// Prior to fixing this logic, this would cause the test to fail/panic.
					//info.Steps()[0] = nil
					labelMap := info.Labels()
					labelMap["foo"] = "bar"
					return ctx, nil
				}).AfterEachFeature(func(ctx context.Context, _ *envconf.Config, info features.Feature) (context.Context, error) {
					val = append(val, "after-each-feature")
					t.Logf("%#v, len(steps)=%v\n", info, len(info.Steps()))
					if info.Labels()["foo"] == "bar" {
						t.Errorf("Expected label from previous feature hook to not include foo:bar")
					}
					return ctx, nil
				})
				f1 := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat-1")
					return ctx
				})
				f2 := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat-2")
					return ctx
				})
				env.Test(t, f1.Feature(), f2.Feature())
				return val
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.setup(t, test.ctx)
			if len(test.expected) != len(result) {
				t.Fatalf("Expected:\n%v but got result:\n%v", test.expected, result)
			}
			for i := range test.expected {
				if result[i] != test.expected[i] {
					t.Errorf("Expected:\n%v but got result:\n%v", test.expected, result)
					break
				}
			}
		})
	}
}

// This test shows the full context propagation from
// environment setup functions (started in main_test.go) down to
// feature step functions.
func TestEnv_Context_Propagation(t *testing.T) {
	f := features.New("test-context-propagation").
		Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
			val, ok := ctx.Value(&ctxTestKeyString{}).([]string)
			if !ok {
				t.Fatal("context value was not int")
			}
			val = append(val, "test-context-propagation")
			return context.WithValue(ctx, &ctxTestKeyString{}, val)
		})

	envForTesting.Test(t, f.Feature())

	env, ok := envForTesting.(*testEnv)
	if !ok {
		t.Fatal("wrong type")
	}

	finalVal, ok := env.ctx.Value(&ctxTestKeyString{}).([]string)
	if !ok {
		t.Fatal("wrong type")
	}

	expected := []string{"setup-1", "setup-2", "before-each-test", "test-context-propagation", "after-each-test"}
	if len(finalVal) != len(expected) {
		t.Fatalf("Expected:\n%v but got result:\n%v", expected, finalVal)
	}
	for i := range finalVal {
		if finalVal[i] != expected[i] {
			t.Errorf("Expected:\n%v but got result:\n%v", expected, finalVal)
			break
		}
	}
}
