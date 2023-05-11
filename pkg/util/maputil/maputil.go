//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package maputil

// Keys returns the slice of keys for the given map.
func Keys[T1 comparable, T2 any](m map[T1]T2) []T1 {
	keys := make([]T1, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// SortKeys returns the slice of keys for the given map, sorted using given
// sorting function.
func SortKeys[T1 comparable, T2 any](m map[T1]T2, sort func([]T1)) []T1 {
	keys := Keys(m)
	sort(keys)
	return keys
}

// Copy returns a shallow copy of the given map.
func Copy[T1 comparable, T2 any](m map[T1]T2) map[T1]T2 {
	newMap := make(map[T1]T2, len(m))
	for k, v := range m {
		newMap[k] = v
	}
	return newMap
}
