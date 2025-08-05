/*
 * Copyright (c) 2025, horoni. All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 * 1. Redistributions of source code must retain the above copyright notice, this
 *    list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright notice,
 *    this list of conditions and the following disclaimer in the documentation
 *    and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
 * ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
 * WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR
 * ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
 * (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
 * LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
 * ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
 * SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sync"
	"time"

	"filippo.io/age"
)

var i int
var i_m sync.Mutex
var wg sync.WaitGroup

func main() {
	var reg_exp string
	var threads_count int

	flag.StringVar(&reg_exp, "r", "age1[^\\s]+", "Regex that key should match to")
	flag.IntVar(&threads_count, "t", 1, "Threads")
	flag.Parse()

	r, _ := regexp.Compile(reg_exp)

	if file_exists("key.txt") {
		log.Fatal("Save and remove key.txt")
		os.Exit(1)
	}

	f, err := os.Create("key.txt")
	if err != nil {
		log.Fatalf("fail to create file: %v", err)
	}
	defer f.Close()

	fmt.Printf("Starting mining with given regex: \"%v\"\n", reg_exp)

	wg.Add(1)

	for j := 0; j < threads_count; j++ {
		go routine_find_key(r, f)
	}
	go routine_dynamic_bar()

	wg.Wait()
}

func routine_find_key(r *regexp.Regexp, f *os.File) {
	defer wg.Done()
	for {
		identity, err := age.GenerateX25519Identity()
		if err != nil {
			log.Fatalf("fail to generate identity: %v", err)
		}

		public_key := identity.Recipient().String()

		if r.MatchString(public_key) == true {
			i_m.Lock()
			fmt.Printf("\nkey(%d) finded: %s\n", i, public_key)
			fmt.Println("Saving and exiting...")
			f.WriteString("# public key: " + public_key +
				"\n" + identity.String() + "\n")
			i_m.Unlock()
			break
		}
		i_m.Lock()
		i++
		i_m.Unlock()
	}
}

func routine_dynamic_bar() {
	start := time.Now()

	for {
		el_tm := time.Since(start)
		fmt.Printf("\rKeys probed: %d, Time elapsed: %s", i, format_duration(el_tm))
		time.Sleep(1 * time.Second)
	}
}

func format_duration(d time.Duration) string {
	d = d.Round(time.Second)

	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour

	hours := d / time.Hour
	d -= hours * time.Hour

	minutes := d / time.Minute
	d -= minutes * time.Minute

	seconds := d / time.Second

	switch {
	case days > 0:
		return fmt.Sprintf("%d days %d h. %d min. %d sec.", days, hours, minutes, seconds)
	case hours > 0:
		return fmt.Sprintf("%d h. %d min. %d sec.", hours, minutes, seconds)
	case minutes > 0:
		return fmt.Sprintf("%d min. %d sec.", minutes, seconds)
	default:
		return fmt.Sprintf("%d sec.", seconds)
	}
}

func file_exists(file_path string) bool {
	_, err := os.Stat(file_path)
	return !os.IsNotExist(err)
}
