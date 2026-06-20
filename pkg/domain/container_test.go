package domain

import "testing"

func TestPrimaryName(t *testing.T) {
	tests := []struct {
		name string
		in   ContainerInfo
		want string
	}{
		{name: "no names", in: ContainerInfo{}, want: ""},
		{name: "single name with leading slash", in: ContainerInfo{Names: []string{"/foo"}}, want: "foo"},
		{name: "single name without slash", in: ContainerInfo{Names: []string{"foo"}}, want: "foo"},
		{name: "multiple names returns first", in: ContainerInfo{Names: []string{"/foo", "/bar"}}, want: "foo"},
		{name: "empty first name", in: ContainerInfo{Names: []string{""}}, want: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.in.PrimaryName()
			if got != tc.want {
				t.Errorf("PrimaryName() = %q, want %q", got, tc.want)
			}
		})
	}
}
