package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path"
	"strconv"
	"strings"
)

type test struct {
	filename                  string
	mfg, size, kv, prop, batt string
	iter                      int

	// summary data
	maxRpm     int
	minVolt    float64
	maxVolt    float64
	maxCurrent float64
	maxThrust  float64
	pwm200     int
	current200 float64
	gpw200     float64
}

func (t *test) summarize() error {
	f, err := os.Open(t.filename)
	if err != nil {
		return err
	}
	defer f.Close()

	c := csv.NewReader(f)
	hdr, err := c.Read()
	if err != nil {
		return err
	}
	idx := map[string]int{}
	for i, n := range hdr {
		idx[n] = i
	}
	floatField := func(s []string, f string) float64 {
		// Error is ignored here -- just return what we find
		v, _ := strconv.ParseFloat(s[idx[f]], 64)
		return v
	}

	t.minVolt = math.MaxFloat64
	has200 := false
	for {
		rec, err := c.Read()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		r, err := strconv.Atoi(rec[idx["rpm"]])
		if err != nil {
			return err
		}

		if r > t.maxRpm {
			t.maxRpm = r
		}
		t.minVolt = math.Min(t.minVolt, floatField(rec, "voltage"))
		t.maxVolt = math.Max(t.maxVolt, floatField(rec, "voltage"))
		t.maxCurrent = math.Max(t.maxCurrent, floatField(rec, "current"))
		t.maxThrust = math.Max(t.maxThrust, floatField(rec, "thrust"))

		if floatField(rec, "thrust") >= 200 && !has200 {
			has200 = true
			t.pwm200, err = strconv.Atoi(rec[idx["pwm"]])
			if err != nil {
				return err
			}
			t.current200 = floatField(rec, "current")
			t.gpw200 = floatField(rec, "gpwatt")
		}
	}
}

func main() {
	tests := []test{}

	for _, fn := range os.Args[1:] {
		parts := strings.Split(path.Base(fn[:len(fn)-4]), "_")
		log.Printf("Doing %v - %v", fn, parts)
		t := test{
			filename: fn,
			mfg:      parts[0],
			size:     parts[1],
			kv:       parts[2][:len(parts[2])-2],
			prop:     parts[3],
			batt:     parts[4],
		}
		if len(parts) > 5 {
			i, err := strconv.Atoi(parts[5])
			if err != nil {
				log.Fatalf("Don't understand iteration in %v: %v", fn, err)
			}
			t.iter = i
		}

		if t.iter == 0 {
			if err := t.summarize(); err != nil {
				log.Fatalf("Error summarizing %#v: %v", t, err)
			}
			log.Printf("Remembering %v: %#v", fn, t)
			tests = append(tests, t)
		}
	}

	fmt.Println(strings.Join([]string{"mfg", "size", "kv", "prop", "batt",
		"maxrpm", "minvolt", "maxvolt", "maxcurr", "maxthr", "pwm200", "curr200", "gpw200",
		"filename"}, ","))

	for _, t := range tests {
		fmt.Printf("%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v\n",
			t.mfg, t.size, t.kv, t.prop, t.batt,
			t.maxRpm, t.minVolt, t.maxVolt, t.maxCurrent, t.maxThrust,
			t.pwm200, t.current200, t.gpw200,
			t.filename)
	}
}
