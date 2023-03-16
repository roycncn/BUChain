package blockchain

import (
	"fmt"
	"github.com/patrickmn/go-cache"
	"strings"
	"testing"
	"time"
)

func TestGenerateUTXO(t *testing.T) {
	strs := []string{"8473671d63b2bfb634d15c6549aa7a9ecfb24846a42c73db23875f164b01302f-0_03c89a1f99b521202cbc702054fef47d323579afb49ffd59a7faed4b60066ea0b4-50", "09a1c85e4a652b10d059bf0c1e341e7d006565da60bb8b57cc9159cd79a7ab5d-0_03c89a1f99b521202cbc702054fef47d323579afb49ffd59a7faed4b60066ea0b4-50", "1b2f3b7d027ab9c5ccb3d5d0fb0529a922c7130840182ec695bc6fa4cfed7453-0_02e90c589d434fe3b70f11bea48bf54922c46646a82d4418d3d9ed16f258a7f88b-50", "273e78783ace7b012d49978fd411d443441e367668270b9b66144548c94b08a9-0_03c89a1f99b521202cbc702054fef47d323579afb49ffd59a7faed4b60066ea0b4-50", "6117d493554032ee2148e0658fcf8c32892ade6c5115fb34e97d0395b159d104-0_039d95813a47234fe889729ff94efa4bb170e4190ba4157f837a68621458018638-50", "0b0cf64030a406382ffd931ea8c6c02fec04ae0ba2f6335b6784d7dbf3e40e1a-0_02e90c589d434fe3b70f11bea48bf54922c46646a82d4418d3d9ed16f258a7f88b-50", "7592a9e5be4848ddd928182824ee1b1b4d8d10d3cbd8cef4d0bd839638b4110f-0_03c89a1f99b521202cbc702054fef47d323579afb49ffd59a7faed4b60066ea0b4-50", "9afe2bbd5a5f87dd2ecc927e93548d0763a7fab38144d2171a73a3f740cbc319-0_03c89a1f99b521202cbc702054fef47d323579afb49ffd59a7faed4b60066ea0b4-50", "e00c7dae7c019a21622f0f2a2ccb0d8044742b1e7bb2efc2bf35cf8d9a665802-0_03c89a1f99b521202cbc702054fef47d323579afb49ffd59a7faed4b60066ea0b4-50", "e3fc4a9007201c5d33ab78de2e0b682dfe874356d2ccd749be61c9b9def62835-0_02e90c589d434fe3b70f11bea48bf54922c46646a82d4418d3d9ed16f258a7f88b-50", "12898a7c25c59c1ae39cec6bbf0058e4e8b6a891af55fcbc105152938308718c-0_03c89a1f99b521202cbc702054fef47d323579afb49ffd59a7faed4b60066ea0b4-50", "a1411bde74069e822d2feb8da0a9c16ff52a4d21de81a9be4e9ebe214854dee7-0_03c89a1f99b521202cbc702054fef47d323579afb49ffd59a7faed4b60066ea0b4-50"}
	newCache := cache.New(10*time.Minute, 10*time.Minute)
	for _, str := range strs {
		arr := strings.Split(str, "_")
		newCache.Set(arr[0], arr[1], 10*time.Minute)
	}

	txi, txo, err := GenerateUTXO("03c89a1f99b521202cbc702054fef47d323579afb49ffd59a7faed4b60066ea0b4", "02e90c589d434fe3b70f11bea48bf54922c46646a82d4418d3d9ed16f258a7f88b", 50, newCache)

	fmt.Println(txi, txo, err)
}