/*
 * Copyright 2020 VMware, Inc.
 * SPDX-License-Identifier: EPL-2.0
 */

package sgtn

import (
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestGetLocaleCompAbnormal(t *testing.T) {

	saved := getDataFromServer
	defer func() { getDataFromServer = saved }()

	errMsg := "TestGetLocaleCompAbnormal"
	getDataFromServer = func(u *url.URL, header map[string]string, data interface{}) (*http.Response, error) {
		return nil, errors.New(errMsg)
	}

	newCfg := testCfg
	newCfg.LocalBundles = ""
	resetInst(&newCfg)

	trans := GetTranslation()

	components, errcomp := trans.GetComponentList(name, version)
	assert.Nil(t, components)
	assert.Contains(t, errcomp.Error(), errMsg)

	components, errcomp = trans.GetComponentList(name, version)
	assert.Nil(t, components)
	assert.Contains(t, errcomp.Error(), errMsg)

	locales, errlocale := trans.GetLocaleList(name, version)
	assert.Nil(t, locales)
	assert.Contains(t, errlocale.Error(), errMsg)

	locales, errlocale = trans.GetLocaleList(name, version)
	assert.Nil(t, locales)
	assert.Contains(t, errlocale.Error(), errMsg)
}

func TestTimeout(t *testing.T) {

	oldClient := httpclient
	defer func() {
		gock.Off()
		httpclient = oldClient
	}()

	newTimeout := time.Microsecond * 10
	transport := http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, newTimeout)
		},
	}
	httpclient = &http.Client{Transport: &transport}

	mockReq := EnableMockDataWithTimes("componentMessages-fr-sunglow", 1)
	mockReq.Mock.Response().Delay(time.Microsecond * 11)

	locale, component := "fr", "sunglow"
	item := &dataItem{dataItemID{itemComponent, name, version, locale, component}, nil, nil}
	item.attrs = getCacheInfo(item)

	resetInst(&testCfg)
	sgtnServer := inst.server

	// Get first time to set server stats as timeout
	err := sgtnServer.Get(item)
	_, ok := errors.Cause(err).(net.Error)
	assert.True(t, true, ok)
	assert.Equal(t, serverTimeout, sgtnServer.status)

	assert.True(t, gock.IsPending())

	// Get second time to get an error "Server times out" immediately
	err = sgtnServer.Get(item)
	assert.Equal(t, "Server times out", err.Error())
}

// Test return to normal status after serverRetryInterval and querying successfully
func TestTimeout2(t *testing.T) {

	defer gock.Off()

	EnableMockDataWithTimes("componentMessages-fr-sunglow", 1)

	locale, component := "fr", "sunglow"
	item := &dataItem{dataItemID{itemComponent, name, version, locale, component}, nil, nil}
	item.attrs = getCacheInfo(item)

	resetInst(&testCfg)
	sgtnServer := inst.server

	sgtnServer.status = serverTimeout
	sgtnServer.lastErrorMoment = time.Now().Unix() - serverRetryInterval - 1
	err := sgtnServer.Get(item)
	assert.Nil(t, err)
	assert.Equal(t, serverNormal, sgtnServer.status)
}