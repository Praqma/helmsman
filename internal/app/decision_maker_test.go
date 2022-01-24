package app

import (
	"reflect"
	"testing"
)

func Test_getValuesFiles(t *testing.T) {
	type args struct {
		r *release
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "test case 1",
			args: args{
				r: &release{
					Name:        "release1",
					Description: "",
					Namespace:   "namespace",
					Enabled:     true,
					Chart:       "repo/chartX",
					Version:     "1.0",
					ValuesFile:  "../../tests/values.yaml",
					Test:        true,
				},
				// s: st,
			},
			want: []string{"-f", "../../tests/values.yaml"},
		},
		{
			name: "test case 2",
			args: args{
				r: &release{
					Name:        "release1",
					Description: "",
					Namespace:   "namespace",
					Enabled:     true,
					Chart:       "repo/chartX",
					Version:     "1.0",
					ValuesFiles: []string{"../../tests/values.yaml"},
					Test:        true,
				},
				// s: st,
			},
			want: []string{"-f", "../../tests/values.yaml"},
		},
		{
			name: "test case 1",
			args: args{
				r: &release{
					Name:        "release1",
					Description: "",
					Namespace:   "namespace",
					Enabled:     true,
					Chart:       "repo/chartX",
					Version:     "1.0",
					ValuesFiles: []string{"../../tests/values.yaml", "../../tests/values2.yaml"},
					Test:        true,
				},
				// s: st,
			},
			want: []string{"-f", "../../tests/values.yaml", "-f", "../../tests/values2.yaml"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.r.getValuesFiles(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getValuesFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_inspectUpgradeScenario(t *testing.T) {
	type args struct {
		r *release
		s *map[string]helmRelease
	}
	tests := []struct {
		name string
		args args
		want decisionType
	}{
		{
			name: "inspectUpgradeScenario() - local chart with different chart name should change",
			args: args{
				r: &release{
					Name:      "release1",
					Namespace: "namespace",
					Version:   "1.0.0",
					Chart:     "./../../tests/chart-test",
					Enabled:   true,
				},
				s: &map[string]helmRelease{
					"release1-namespace": {
						Namespace: "namespace",
						Chart:     "chart-1.0.0",
					},
				},
			},
			want: change,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outcome := plan{}
			cs := currentState{releases: *tt.args.s}

			// Act
			c, _ := getChartInfo(tt.args.r.Chart, tt.args.r.Version)
			if err := cs.inspectUpgradeScenario(tt.args.r, &outcome, c); err != nil {
				t.Errorf("inspectUpgradeScenario() error = %v", err)
			}
			got := outcome.Decisions[0].Type
			t.Log(outcome.Decisions[0].Description)

			// Assert
			if got != tt.want {
				t.Errorf("decide() = %s, want %s", got, tt.want)
			}
		})
	}
}

func Test_decide(t *testing.T) {
	type args struct {
		r string
		s *state
	}
	tests := []struct {
		name       string
		targetFlag []string
		args       args
		want       decisionType
	}{
		{
			name:       "decide() - targetMap does not contain this service - skip",
			targetFlag: []string{"someOtherRelease"},
			args: args{
				r: "release1",
				s: &state{
					Apps: map[string]*release{
						"release1": {
							Name:      "release1",
							Namespace: "namespace",
							Enabled:   true,
						},
					},
				},
			},
			want: ignored,
		},
		{
			name:       "decide() - targetMap does not contain this service either - skip",
			targetFlag: []string{"someOtherRelease", "norThisOne"},
			args: args{
				r: "release1",
				s: &state{
					Apps: map[string]*release{
						"release1": {
							Name:      "release1",
							Namespace: "namespace",
							Enabled:   true,
						},
					},
				},
			},
			want: ignored,
		},
		{
			name:       "decide() - targetMap is empty - will install",
			targetFlag: []string{},
			args: args{
				r: "release4",
				s: &state{
					Apps: map[string]*release{
						"release4": {
							Name:      "release4",
							Namespace: "namespace",
							Enabled:   true,
						},
					},
				},
			},
			want: create,
		},
		{
			name:       "decide() - targetMap is exactly this service - will install",
			targetFlag: []string{"thisRelease"},
			args: args{
				r: "thisRelease",
				s: &state{
					Apps: map[string]*release{
						"thisRelease": {
							Name:      "thisRelease",
							Namespace: "namespace",
							Enabled:   true,
						},
					},
				},
			},
			want: create,
		},
		{
			name:       "decide() - targetMap contains this service - will install",
			targetFlag: []string{"notThisOne", "thisRelease"},
			args: args{
				r: "thisRelease",
				s: &state{
					Apps: map[string]*release{
						"thisRelease": {
							Name:      "thisRelease",
							Namespace: "namespace",
							Enabled:   true,
						},
					},
				},
			},
			want: create,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := newCurrentState()
			tt.args.s.disableUntargetedApps([]string{}, tt.targetFlag)
			settings := config{}
			outcome := plan{}
			// Act
			cs.decide(tt.args.s.Apps[tt.args.r], tt.args.s.Namespaces[tt.args.s.Apps[tt.args.r].Namespace], &outcome, &chartInfo{}, settings)
			got := outcome.Decisions[0].Type
			t.Log(outcome.Decisions[0].Description)

			// Assert
			if got != tt.want {
				t.Errorf("decide() = %s, want %s", got, tt.want)
			}
		})
	}
}

func Test_decide_skip_ignored_apps(t *testing.T) {
	type args struct {
		rs []string
		s  *state
	}
	tests := []struct {
		name       string
		targetFlag []string
		args       args
		want       int
	}{
		{
			name:       "decide() - skipIgnoredApps is defined and target flag contains 2 app names - plan should have 2 decisions",
			targetFlag: []string{"service1", "service2"},
			args: args{
				rs: []string{"service1", "service2"},
				s: &state{
					Apps: map[string]*release{
						"service1": {
							Name:      "service1",
							Namespace: "namespace",
							Enabled:   true,
						},
						"service2": {
							Name:      "service2",
							Namespace: "namespace",
							Enabled:   true,
						},
					},
				},
			},
			want: 2,
		},
		{
			name:       "decide() - skipIgnoredApps is defined and target flag contains just 1 app name - plan should have 1 decision - one app should be ignored",
			targetFlag: []string{"service1"},
			args: args{
				rs: []string{"service1", "service2"},
				s: &state{
					Apps: map[string]*release{
						"service1": {
							Name:      "service1",
							Namespace: "namespace",
							Enabled:   true,
						},
						"service2": {
							Name:      "service2",
							Namespace: "namespace",
							Enabled:   true,
						},
					},
				},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := newCurrentState()
			tt.args.s.disableUntargetedApps([]string{}, tt.targetFlag)
			settings := config{
				SkipIgnoredApps: true,
			}
			outcome := plan{}
			// Act
			for _, r := range tt.args.rs {
				cs.decide(tt.args.s.Apps[r], tt.args.s.Namespaces[tt.args.s.Apps[r].Namespace], &outcome, &chartInfo{}, settings)
			}
			got := outcome.Decisions
			t.Log(outcome.Decisions)

			// Assert
			if len(got) != tt.want {
				t.Errorf("decide() = %d, want %d", len(got), tt.want)
			}
		})
	}
}

func Test_decide_group(t *testing.T) {
	type args struct {
		s *state
	}
	tests := []struct {
		name      string
		groupFlag []string
		args      args
		want      map[string]bool
	}{
		{
			name:      "decide() - groupMap does not contain this service - skip",
			groupFlag: []string{"some-group"},
			args: args{
				s: &state{
					Apps: map[string]*release{
						"release1": {
							Name:      "release1",
							Namespace: "namespace",
							Group:     "run-me-not",
							Enabled:   true,
						},
					},
				},
			},
			want: map[string]bool{},
		},
		{
			name:      "decide() - groupMap contains this service - proceed",
			groupFlag: []string{"run-me"},
			args: args{
				s: &state{
					Apps: map[string]*release{
						"release1": {
							Name:      "release1",
							Namespace: "namespace",
							Group:     "run-me",
							Enabled:   true,
						},
						"release2": {
							Name:      "release2",
							Namespace: "namespace",
							Group:     "run-me-not",
							Enabled:   true,
						},
						"release3": {
							Name:      "release3",
							Namespace: "namespace2",
							Group:     "run-me-not",
							Enabled:   true,
						},
					},
				},
			},
			want: map[string]bool{
				"release1": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.s.disableUntargetedApps(tt.groupFlag, []string{})
			if len(tt.args.s.TargetMap) != len(tt.want) {
				t.Errorf("decide() = %d, want %d", len(tt.args.s.TargetMap), len(tt.want))
			}
		})
	}
}

// String allows for pretty printing decisionType const
func (dt decisionType) String() string {
	switch dt {
	case create:
		return "create"
	case change:
		return "change"
	case remove:
		return "remove"
	case noop:
		return "noop"
	}
	return "unknown"
}
