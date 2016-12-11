package msgtypes

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"strconv"
)

// -----------------------------------------------------------------------------
// Uuid4() generates a totally random UUID (version 4) as described in
// RFC 4122.
// Copyright (c) 2016, The Gocql authors.
//
// This code is taken from https://github.com/gocql/gocql.
// Please refer to https://github.com/gocql/gocql/blob/master/LICENSE.
//
func UUID4() (*UUID, error) {
	var u UUID
	_, err := io.ReadFull(rand.Reader, u[:])
	if err != nil {
		return &u, err
	}
	u[6] &= 0x0F // clear version
	u[6] |= 0x40 // set version to 4 (random uuid)
	u[8] &= 0x3F // clear variant
	u[8] |= 0x80 // set to IETF variant
	return &u, nil
}

// -----------------------------------------------------------------------------
// Accompanying type UUID to the Uuid4 function. Contains some convenience
// methods like parsing and formatting.
//
func UUID4Empty() *UUID {
	var uuid UUID
	return &(uuid)
}

type UUID [16]byte

func (this *UUID) IsValid() bool {
	if this == nil {
		panic("uuid nil pointer exception")
	}
	// The only requirement for the UUID is to have the corresponding flags set
	// correctly.
	return this[6]&0x40 != 0 && this[8]&0x80 != 0
}

func (this *UUID) ToString() string {
	if this == nil {
		panic("uuid nil pointer exception")
	}
	return (hex.EncodeToString(this[0:4]) + "-" +
		hex.EncodeToString(this[4:6]) + "-" +
		hex.EncodeToString(this[6:8]) + "-" +
		hex.EncodeToString(this[8:10]) + "-" +
		hex.EncodeToString(this[10:16]))
}

func (this *UUID) FromString(uuid string) (err error) {
	if this == nil {
		panic("uuid nil pointer exception")
	}
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	var (
		n int
		e error
	)
	if len(uuid) != 36 {
		return errors.New("Expected UUID of format xxxxxxxx-xxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx (hex)")
	}
	bytes := []byte(uuid)
	n, e = hex.Decode(this[0:4], bytes[0:8])
	this.checkParseStep(n, 4, e)
	n, e = hex.Decode(this[4:6], bytes[9:13])
	this.checkParseStep(n, 2, e)
	n, e = hex.Decode(this[6:8], bytes[14:18])
	this.checkParseStep(n, 2, e)
	n, e = hex.Decode(this[8:10], bytes[19:23])
	this.checkParseStep(n, 2, e)
	n, e = hex.Decode(this[10:16], bytes[24:36])
	this.checkParseStep(n, 6, e)
	return
}

func (this *UUID) checkParseStep(n int, c int, e error) {
	if e != nil {
		panic(e)
	}
	if n != c {
		panic(errors.New("Expected " + strconv.FormatInt(int64(c), 10) + " hex-encoded bytes, but got only " + strconv.FormatInt(int64(n), 10)))
	}
}

func (this *UUID) ToBytes() []byte {
	if this == nil {
		panic("uuid nil pointer exception")
	}
	return this[:]
}

func (this *UUID) FromBytes(bytes []byte) error {
	if this == nil {
		panic("uuid nil pointer exception")
	}
	if len(bytes) != 16 {
		return errors.New("Invalid UUID size")
	}
	copy(this[:], bytes)
	return nil
}

func UUIDFromBytes(bytes []byte) (*UUID, error) {
	uuid := UUID4Empty()
	err := uuid.FromBytes(bytes)
	return uuid, err
}

func UUIDFromString(str string) (*UUID, error) {
	uuid := UUID4Empty()
	err := uuid.FromString(str)
	return uuid, err
}
