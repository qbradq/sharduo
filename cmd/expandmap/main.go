package main

import (
	"fmt"
	"log"
	"os"
)

const (
	oldMapWidth         int = 6144
	newMapWidth         int = 7168
	expandedWidth       int = newMapWidth - oldMapWidth
	mapHeight           int = 4096
	chunkWidth          int = 8
	chunkHeight         int = 8
	expandedWidthChunks int = expandedWidth / chunkWidth
	mapHeightChunks     int = mapHeight / chunkHeight
)

func help() {
	fmt.Println(
		"Usage: expandmap\n" +
			"\n" +
			"    Expands a 6144x4096 Felucca / Trammel map into a 7168x4096 map. Execute\n" +
			"this program inside the Ultima Online directory containing the files you wish\n" +
			"to modify. This program modifies map0.mul and staidx0.mul. The original\n" +
			"files will be backed up as *.mul.bak.")
}

func main() {
	if len(os.Args) != 1 {
		help()
		os.Exit(1)
	}
	// Expand terrain map with empty tiles (the black stuff around dungeons)
	blockHeaderBlob := []byte{0, 0, 0, 0}
	blankTileBlob := []byte{0x44, 0x02, 0x00}
	mapData, err := os.ReadFile("map0.mul")
	if err != nil {
		help()
		log.Fatal(err)
	}
	if err := os.Rename("map0.mul", "map0.mul.bak"); err != nil {
		help()
		log.Fatal(err)
	}
	o, err := os.Create("map0.mul")
	if err != nil {
		help()
		log.Fatal(err)
	}
	o.Write(mapData)
	for iBlock := 0; iBlock < expandedWidthChunks*mapHeightChunks; iBlock++ {
		o.Write(blockHeaderBlob)
		for iTile := 0; iTile < chunkWidth*chunkHeight; iTile++ {
			o.Write(blankTileBlob)
		}
	}
	o.Close()
	// Expand statics index with empty sectors
	blankStaticsBlob := []byte{0xff, 0xff, 0xff, 0xff, 0, 0, 0, 0, 0, 0, 0, 0}
	staticsData, err := os.ReadFile("staidx0.mul")
	if err != nil {
		help()
		log.Fatal(err)
	}
	if err := os.Rename("staidx0.mul", "staidx0.mul.bak"); err != nil {
		help()
		log.Fatal(err)
	}
	o, err = os.Create("staidx0.mul")
	if err != nil {
		help()
		log.Fatal(err)
	}
	o.Write(staticsData)
	for iBlock := 0; iBlock < expandedWidthChunks*mapHeightChunks; iBlock++ {
		o.Write(blankStaticsBlob)
	}
	o.Close()
}
