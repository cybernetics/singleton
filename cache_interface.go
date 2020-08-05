/*
 * Copyright 2020 VMware, Inc.
 * SPDX-License-Identifier: EPL-2.0
 */

package sgtn

// Cache the interface of Cache
type Cache interface {
	// Get
	Get(key interface{}) (value interface{}, found bool)

	// Set
	Set(key interface{}, value interface{})
}