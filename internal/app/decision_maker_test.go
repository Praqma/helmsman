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
				//s: st,
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
				//s: st,
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
				//s: st,
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
			chartName := extractChartName(tt.args.r.Chart, tt.args.r.Version)
			chartVersion, _ := getChartVersion(tt.args.r.Chart, tt.args.r.Version)
			cs.inspectUpgradeScenario(tt.args.r, &outcome, chartName, chartVersion)
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
		r *release
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
				r: &release{
					Name:      "release1",
					Namespace: "namespace",
					Enabled:   true,
				},
				s: &state{},
			},
			want: ignored,
		},
		{
			name:       "decide() - targetMap does not contain this service - skip",
			targetFlag: []string{"someOtherRelease", "norThisOne"},
			args: args{
				r: &release{
					Name:      "release1",
					Namespace: "namespace",
					Enabled:   true,
				},
				s: &state{},
			},
			want: ignored,
		},
		{
			name:       "decide() - targetMap is empty - will install",
			targetFlag: []string{},
			args: args{
				r: &release{
					Name:      "release4",
					Namespace: "namespace",
					Enabled:   true,
				},
				s: &state{},
			},
			want: create,
		},
		{
			name:       "decide() - targetMap is exactly this service - will install",
			targetFlag: []string{"thisRelease"},
			args: args{
				r: &release{
					Name:      "thisRelease",
					Namespace: "namespace",
					Enabled:   true,
				},
				s: &state{},
			},
			want: create,
		},
		{
			name:       "decide() - targetMap contains this service - will install",
			targetFlag: []string{"notThisOne", "thisRelease"},
			args: args{
				r: &release{
					Name:      "thisRelease",
					Namespace: "namespace",
					Enabled:   true,
				},
				s: &state{},
			},
			want: create,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := newCurrentState()
			tt.args.s.TargetMap = make(map[string]bool)

			for _, target := range tt.targetFlag {
				tt.args.s.TargetMap[target] = true
			}
			outcome := plan{}
			// Act
			cs.decide(tt.args.r, tt.args.s, &outcome, "", "")
			got := outcome.Decisions[0].Type
			t.Log(outcome.Decisions[0].Description)

			// Assert
			if got != tt.want {
				t.Errorf("decide() = %s, want %s", got, tt.want)
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
							Group:     "run-me",
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
			tt.args.s.GroupMap = make(map[string]bool)
			for _, group := range tt.groupFlag {
				tt.args.s.GroupMap[group] = true
			}
			tt.args.s.TargetMap = tt.args.s.getAppsInGroupsAsTargetMap()
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
	case delete:
		return "delete"
	case noop:
		return "noop"
	}
	return "unknown"
}
