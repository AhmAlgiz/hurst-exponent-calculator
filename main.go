package main

import (
	"encoding/csv"
	"fmt"
	"image/color"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/gonum/stat"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
)

func readCSV(filename string) ([]float64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	csvReader.Comma = ';'

	res := make([]float64, 0)

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			return res, nil
		}
		if err != nil {
			return nil, err
		}
		record[1] = strings.Replace(record[1], ",", ".", 1)

		value, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			return nil, err
		}

		res = append(res, value)
	}
}

func plotGraph(x, y []float64, text string) error {
	p := plot.New()

	f, err := os.Create(fmt.Sprintf("%s.png", text))
	if err != nil {
		return fmt.Errorf("error creating png: %v", err)
	}
	defer f.Close()

	pxys := make(plotter.XYs, len(x))
	for i := range x {
		pxys[i].X = x[i]
		pxys[i].Y = y[i]
	}

	s, err := plotter.NewScatter(pxys)
	if err != nil {
		return fmt.Errorf("error creating scatter: %v", err)
	}
	s.Color = color.RGBA{R: 255, A: 255}

	p.Add(s)

	l, err := plotter.NewLine(pxys)
	if err != nil {
		return fmt.Errorf("error creating lines: %v", err)
	}
	l.Color = color.RGBA{G: 255, A: 255}

	p.Add(l)

	wt, err := p.WriterTo(512, 512, "png")
	if err != nil {
		return fmt.Errorf("error init plot writer: %v", err)
	}

	_, err = wt.WriteTo(f)
	if err != nil {
		return fmt.Errorf("error writting plot: %v", err)
	}

	return nil
}

func hcalc(data []float64, text string) float64 {

	pmax := 10
	e := make([]float64, pmax)
	n := make([]float64, pmax)
	w := make([]float64, pmax)
	xw := make([]float64, len(data))
	for i := range xw {
		xw[i] = 1
	}

	for p := 0; p < pmax; p++ {
		dlen := len(data) - pmax
		dx := make([]float64, dlen)
		for i := 0; i < dlen; i++ {
			dx[i] = data[i+p+1] - data[i]
		}
		e[p] = math.Log(float64(p + 1))
		n[p] = math.Log(stat.StdDev(dx, xw[:len(data)-pmax]))
		w[p] = 1
	}

	a, b := stat.LinearRegression(e, n, w, false)

	fx := make([]float64, pmax)
	for i := range e {
		fx[i] = a + b*e[i]
	}

	if err := plotGraph(e, n, fmt.Sprintf("log(std)-log(p)_%v", text)); err != nil {
		log.Fatalf("error plotting: %v", err)
	}

	if err := plotGraph(e, fx, fmt.Sprintf("fx-log(p)_%v", text)); err != nil {
		log.Fatalf("error plotting: %v", err)
	}

	return b
}

func main() {
	data, err := readCSV("USD.csv")
	if err != nil {
		log.Fatalf("Ошибка при чтении файла: %v", err)
	}

	x := make([]float64, len(data))
	for i := range x {
		x[i] = float64(i)
	}

	h := hcalc(data, "USD")
	fmt.Printf("Показатель Херста для USD равен: %.2f \n", h)

	if err := plotGraph(x, data, fmt.Sprintf("%v", "USD")); err != nil {
		log.Fatalf("error plotting: %v", err)
	}

	data, err = readCSV("JPY.csv")
	if err != nil {
		log.Fatalf("Ошибка при чтении файла: %v", err)
	}
	h = hcalc(data, "JPY")
	fmt.Printf("Показатель Херста для JPY равен: %.2f", h)

	if err := plotGraph(x, data, fmt.Sprintf("%v", "JPY")); err != nil {
		log.Fatalf("error plotting: %v", err)
	}

}
