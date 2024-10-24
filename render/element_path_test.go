package render

import "testing"

func TestNodePath_MutexWith(t *testing.T) {
	type args struct {
		p2 ElementPath
	}
	tests := []struct {
		name string
		p    ElementPath
		args args
		want bool
	}{
		{
			name: "c1/c4/s1 & c1/c4/s1/c5",
			p: func() ElementPath {
				var p ElementPath
				p.AddNode(1, true, false)
				p.AddNode(4, true, false)
				p.AddNode(1, false, false)
				return p
			}(),
			args: args{
				p2: func() ElementPath {
					var p ElementPath
					p.AddNode(1, true, false)
					p.AddNode(4, true, false)
					p.AddNode(1, false, false)
					p.AddNode(5, true, false)
					return p
				}(),
			},
			want: false,
		},
		{
			name: "c1/c4/s1/c5 & c1/c4/s1/c5",
			p: func() ElementPath {
				var p ElementPath
				p.AddNode(1, true, false)
				p.AddNode(4, true, false)
				p.AddNode(1, false, false)
				p.AddNode(5, true, false)
				return p
			}(),
			args: args{
				p2: func() ElementPath {
					var p ElementPath
					p.AddNode(1, true, false)
					p.AddNode(4, true, false)
					p.AddNode(1, false, false)
					p.AddNode(5, true, false)
					return p
				}(),
			},
			want: true,
		},
		{
			name: "c1/c4 & c1/c4/s1/c5",
			p: func() ElementPath {
				var p ElementPath
				p.AddNode(1, true, false)
				p.AddNode(4, true, false)
				return p
			}(),
			args: args{
				p2: func() ElementPath {
					var p ElementPath
					p.AddNode(1, true, false)
					p.AddNode(4, true, false)
					p.AddNode(1, false, false)
					p.AddNode(5, true, false)
					return p
				}(),
			},
			want: true,
		},
		{
			name: "c1/c2 & c1/c4/s1/c5",
			p: func() ElementPath {
				var p ElementPath
				p.AddNode(1, true, false)
				p.AddNode(2, true, false)
				return p
			}(),
			args: args{
				p2: func() ElementPath {
					var p ElementPath
					p.AddNode(1, true, false)
					p.AddNode(4, true, false)
					p.AddNode(1, false, false)
					p.AddNode(5, true, false)
					return p
				}(),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.MutexWith(tt.args.p2); got != tt.want {
				t.Errorf("NodePath.MutexWith() = %v, want %v", got, tt.want)
			}
		})
	}
}
