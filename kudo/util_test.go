package main


import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestId(t *testing.T){
	inName := "foo"
	inNamespace := "bar"
	id := id(inName, inNamespace)
	assert.Equal(t,id,"foo_bar")
	outName, outNamespace, err := idParts(id)
	assert.Nil(t,err)
	assert.Equal(t,inName, outName)
	assert.Equal(t, inNamespace, outNamespace)
}