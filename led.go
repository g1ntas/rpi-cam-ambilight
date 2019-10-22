package main

import (
	"fmt"
	"github.com/stianeikeland/go-rpio"
	"time"
)

type WS2801Led struct{
	State []uint8
	Count int
	spi rpio.SpiDev
}

func NewWS2801Led(dev rpio.SpiDev, amountOfLeds int) (*WS2801Led, error) {
	if amountOfLeds <= 0 {
		return nil, fmt.Errorf("amount of leds should be greater than zero")
	}
	if err := rpio.Open(); err != nil {
		return nil, err
	}
	led := &WS2801Led{}
	if err := rpio.SpiBegin(dev); err != nil {
		return nil, err
	}
	led.Count = amountOfLeds
	led.State = make([]uint8, led.Count*3)
	led.spi = dev
	rpio.SpiSpeed(1000000) // 1 mHZ
	rpio.SpiChipSelect(0)
	return led, nil
}

func (led *WS2801Led) Close() error {
	rpio.SpiEnd(led.spi)
	err := rpio.Close()
	return err
}

func (led *WS2801Led) UpdatePixel(i int, r, g, b uint8) error {
	if i < 0 || i >= led.Count {
		return fmt.Errorf("LED index %d is out of range (0-%d)", i, led.Count)
	}
	led.State[i*3] = r
	led.State[i*3+1] = b
	led.State[i*3+2] = g
	led.Update()
	return nil
}

func (led *WS2801Led) Update() {
	rpio.SpiTransmit(led.State...)
	time.Sleep(2 * time.Millisecond)
}
