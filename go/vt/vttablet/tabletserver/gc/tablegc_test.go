/*
Copyright 2020 The Vitess Authors.

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

package gc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNextTableToPurge(t *testing.T) {
	tt := []struct {
		tables []string
		next   string
		ok     bool
	}{
		{
			tables: []string{},
			ok:     false,
		},
		{
			tables: []string{
				"_vt_PURGE_6ace8bcef73211ea87e9f875a4d24e90_20200915120410",
				"_vt_PURGE_2ace8bcef73211ea87e9f875a4d24e90_20200915120411",
				"_vt_PURGE_3ace8bcef73211ea87e9f875a4d24e90_20200915120412",
				"_vt_PURGE_4ace8bcef73211ea87e9f875a4d24e90_20200915120413",
			},
			next: "_vt_PURGE_6ace8bcef73211ea87e9f875a4d24e90_20200915120410",
			ok:   true,
		},
		{
			tables: []string{
				"_vt_PURGE_2ace8bcef73211ea87e9f875a4d24e90_20200915120411",
				"_vt_PURGE_3ace8bcef73211ea87e9f875a4d24e90_20200915120412",
				"_vt_PURGE_6ace8bcef73211ea87e9f875a4d24e90_20200915120410",
				"_vt_PURGE_4ace8bcef73211ea87e9f875a4d24e90_20200915120413",
			},
			next: "_vt_PURGE_6ace8bcef73211ea87e9f875a4d24e90_20200915120410",
			ok:   true,
		},
	}
	for _, ts := range tt {
		collector := &TableGC{
			purgingTables: make(map[string]bool),
		}
		for _, table := range ts.tables {
			collector.purgingTables[table] = true
		}
		next, ok := collector.nextTableToPurge()
		assert.Equal(t, ts.ok, ok)
		if ok {
			assert.Equal(t, ts.next, next)
		}
	}
}
