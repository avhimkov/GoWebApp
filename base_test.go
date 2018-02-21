package main

import "testing"

func TestPerson_save(t *testing.T) {
	tests := []struct {
		name    string
		p       *Person
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.save(); (err != nil) != tt.wantErr {
				t.Errorf("Person.save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
