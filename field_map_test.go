// Copyright (c) quickfixengine.org  All rights reserved.
//
// This file may be distributed under the terms of the quickfixengine.org
// license as defined by quickfixengine.org and appearing in the file
// LICENSE included in the packaging of this file.
//
// This file is provided AS IS with NO WARRANTY OF ANY KIND, INCLUDING
// THE WARRANTY OF DESIGN, MERCHANTABILITY AND FITNESS FOR A
// PARTICULAR PURPOSE.
//
// See http://www.quickfixengine.org/LICENSE for licensing information.
//
// Contact ask@quickfixengine.org if any conditions of this licensing
// are not clear to you.

package quickfix

import (
	"bytes"
	"testing"

	"github.com/karlseguin/jsonwriter"
	"github.com/quickfixgo/quickfix/datadictionary"
	"github.com/stretchr/testify/assert"
)

func TestFieldMap_Clear(t *testing.T) {
	var fMap FieldMap
	fMap.init()

	fMap.SetField(1, FIXString("hello"))
	fMap.SetField(2, FIXString("world"))

	fMap.Clear()

	if fMap.Has(1) || fMap.Has(2) {
		t.Error("All fields should be cleared")
	}
}

func TestFieldMap_SetAndGet(t *testing.T) {
	var fMap FieldMap
	fMap.init()

	fMap.SetField(1, FIXString("hello"))
	fMap.SetField(2, FIXString("world"))

	testCases := []struct {
		tag         Tag
		expectErr   bool
		expectValue string
	}{
		{tag: 1, expectValue: "hello"},
		{tag: 2, expectValue: "world"},
		{tag: 44, expectErr: true},
	}

	for _, tc := range testCases {
		var testField FIXString
		err := fMap.GetField(tc.tag, &testField)

		if tc.expectErr {
			assert.NotNil(t, err, "Expected Error")
			continue
		}

		assert.Nil(t, err, "Unexpected error")
		assert.Equal(t, tc.expectValue, string(testField))
	}
}

func TestFieldMap_Length(t *testing.T) {
	var fMap FieldMap
	fMap.init()
	fMap.SetField(1, FIXString("hello"))
	fMap.SetField(2, FIXString("world"))
	fMap.SetField(8, FIXString("FIX.4.4"))
	fMap.SetField(9, FIXInt(100))
	fMap.SetField(10, FIXString("100"))
	assert.Equal(t, 16, fMap.length(), "Length should include all fields but beginString, bodyLength, and checkSum")
}

func TestFieldMap_Total(t *testing.T) {
	var fMap FieldMap
	fMap.init()
	fMap.SetField(1, FIXString("hello"))
	fMap.SetField(2, FIXString("world"))
	fMap.SetField(8, FIXString("FIX.4.4"))
	fMap.SetField(Tag(9), FIXInt(100))
	fMap.SetField(10, FIXString("100"))

	assert.Equal(t, 2116, fMap.total(), "Total should includes all fields but checkSum")
}

func TestFieldMap_TypedSetAndGet(t *testing.T) {
	var fMap FieldMap
	fMap.init()

	fMap.SetString(1, "hello")
	fMap.SetInt(2, 256)

	s, err := fMap.GetString(1)
	assert.Nil(t, err)
	assert.Equal(t, "hello", s)

	i, err := fMap.GetInt(2)
	assert.Nil(t, err)
	assert.Equal(t, 256, i)

	_, err = fMap.GetInt(1)
	assert.NotNil(t, err, "Type mismatch should occur error")

	s, err = fMap.GetString(2)
	assert.Nil(t, err)
	assert.Equal(t, "256", s)

	b, err := fMap.GetBytes(1)
	assert.Nil(t, err)
	assert.True(t, bytes.Equal([]byte("hello"), b))
}

func TestFieldMap_BoolTypedSetAndGet(t *testing.T) {
	var fMap FieldMap
	fMap.init()

	fMap.SetBool(1, true)
	v, err := fMap.GetBool(1)
	assert.Nil(t, err)
	assert.True(t, v)

	s, err := fMap.GetString(1)
	assert.Nil(t, err)
	assert.Equal(t, "Y", s)

	fMap.SetBool(2, false)
	v, err = fMap.GetBool(2)
	assert.Nil(t, err)
	assert.False(t, v)

	s, err = fMap.GetString(2)
	assert.Nil(t, err)
	assert.Equal(t, "N", s)
}

func TestFieldMap_CopyInto(t *testing.T) {
	var fMapA FieldMap
	fMapA.initWithOrdering(headerFieldOrdering)
	fMapA.SetString(9, "length")
	fMapA.SetString(8, "begin")
	fMapA.SetString(35, "msgtype")
	fMapA.SetString(1, "a")
	assert.Equal(t, []Tag{8, 9, 35, 1}, fMapA.sortedTags())

	var fMapB FieldMap
	fMapB.init()
	fMapB.SetString(1, "A")
	fMapB.SetString(3, "C")
	fMapB.SetString(4, "D")
	assert.Equal(t, fMapB.sortedTags(), []Tag{1, 3, 4})

	fMapA.CopyInto(&fMapB)

	assert.Equal(t, []Tag{8, 9, 35, 1}, fMapB.sortedTags())

	// new fields
	s, err := fMapB.GetString(35)
	assert.Nil(t, err)
	assert.Equal(t, "msgtype", s)

	// existing fields overwritten
	s, err = fMapB.GetString(1)
	assert.Nil(t, err)
	assert.Equal(t, "a", s)

	// old fields cleared
	_, err = fMapB.GetString(3)
	assert.NotNil(t, err)

	// check that ordering is overwritten
	fMapB.SetString(2, "B")
	assert.Equal(t, []Tag{8, 9, 35, 1, 2}, fMapB.sortedTags())

	// updating the existing map doesn't affect the new
	fMapA.init()
	fMapA.SetString(1, "AA")
	s, err = fMapB.GetString(1)
	assert.Nil(t, err)
	assert.Equal(t, "a", s)
	fMapA.Clear()
	s, err = fMapB.GetString(1)
	assert.Nil(t, err)
	assert.Equal(t, "a", s)
}

func TestFieldMap_Remove(t *testing.T) {
	var fMap FieldMap
	fMap.init()

	fMap.SetField(1, FIXString("hello"))
	fMap.SetField(2, FIXString("world"))

	fMap.Remove(1)
	assert.False(t, fMap.Has(1))
	assert.True(t, fMap.Has(2))
}

func TestFieldMap_ToJSON(t *testing.T) {
	// Load FIX44 data dictionary for field type information
	dict, err := datadictionary.Parse("spec/FIX44.xml")
	if err != nil {
		t.Fatalf("Failed to parse data dictionary: %v", err)
	}

	var fMap FieldMap
	fMap.init()

	// Add at least 10 regular fields
	fMap.SetString(1, "ACCOUNT123")         // Account
	fMap.SetString(11, "ORDER-001")         // ClOrdID
	fMap.SetString(21, "1")                 // HandlInst
	fMap.SetString(38, "100")               // OrderQty
	fMap.SetString(40, "2")                 // OrdType (Limit)
	fMap.SetString(44, "100.50")            // Price
	fMap.SetString(54, "1")                 // Side (Buy)
	fMap.SetString(55, "AAPL")              // Symbol
	fMap.SetString(59, "0")                 // TimeInForce (Day)
	fMap.SetString(60, "20230101-10:00:00") // TransactTime

	// Create groups with at least 3 fields each
	// Group 1: NoPartyIDs (453) with PartyID(448), PartyIDSource(447), PartyRole(452)
	partyGroup1 := NewRepeatingGroup(453, GroupTemplate{
		GroupElement(448), // PartyID
		GroupElement(447), // PartyIDSource
		GroupElement(452), // PartyRole
	})
	partyGroup1.Add().SetString(448, "BROKER1").SetString(447, "D").SetString(452, "1")
	partyGroup1.Add().SetString(448, "CLIENT1").SetString(447, "D").SetString(452, "2")
	fMap.SetGroup(partyGroup1)

	// Group 2: NoTradingSessions (386) with TradingSessionID(336), TradingSessionSubID(625), TradSesMethod(340)
	sessionGroup := NewRepeatingGroup(386, GroupTemplate{
		GroupElement(336), // TradingSessionID
		GroupElement(625), // TradingSessionSubID
		GroupElement(340), // TradSesMethod
	})
	sessionGroup.Add().SetString(336, "SESSION1").SetString(625, "SUB1").SetString(340, "1")
	sessionGroup.Add().SetString(336, "SESSION2").SetString(625, "SUB2").SetString(340, "2")
	fMap.SetGroup(sessionGroup)

	// Convert to JSON
	buf := new(bytes.Buffer)
	w := jsonwriter.New(buf)
	w.RootObject(func() {
		err = fMap.ToJSON(w, true, dict)
	})

	assert.Nil(t, err)

	jsonBytes := buf.Bytes()

	// Verify JSON is not empty
	if len(jsonBytes) == 0 {
		t.Error("JSON output should not be empty")
	}

	// Basic validation that JSON contains expected fields
	jsonStr := string(jsonBytes)
	t.Logf("Generated JSON: %s", jsonStr)

	// Check that some key fields are present in the JSON
	assert.Contains(t, jsonStr, "Account")
	assert.Contains(t, jsonStr, "ClOrdID")
	assert.Contains(t, jsonStr, "Symbol")
	assert.Contains(t, jsonStr, "NoPartyIDs")
	assert.Contains(t, jsonStr, "NoTradingSessions")
}
