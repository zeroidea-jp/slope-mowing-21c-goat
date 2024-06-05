package main

import (
	"fmt"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/devices/v3/ledc"
	"periph.io/x/host/v3"
)

const (
	pwmPin   = "GPIO18"          // PWM信号のピン番号
	telcoPin = "GPIO23"          // テレメトリー信号のピン番号
	freq     = 50 * physic.Hertz // PWM周波数
	dutyMin  = 0.05              // 最小デューティ比
	dutyMax  = 0.1               // 最大デューティ比
)

func main() {
	// ホストの初期化
	if _, err := host.Init(); err != nil {
		fmt.Printf("failed to initialize host: %v\n", err)
		return
	}

	// PWMピンの設定
	pwm := gpioreg.ByName(pwmPin)
	if pwm == nil {
		fmt.Printf("failed to find PWM pin %s\n", pwmPin)
		return
	}

	// LEDCドライバの初期化
	l := ledc.New(pwm, freq, dutyMax, dutyMin)
	if err := l.Start(); err != nil {
		fmt.Printf("failed to start LEDC driver: %v\n", err)
		return
	}
	defer l.Stop()

	// テレメトリーピンの設定
	telco := gpioreg.ByName(telcoPin)
	if telco == nil {
		fmt.Printf("failed to find telemetry pin %s\n", telcoPin)
		return
	}
	if err := telco.In(gpio.PullNoChange, gpio.RisingEdge); err != nil {
		fmt.Printf("failed to set telemetry pin as input: %v\n", err)
		return
	}

	// モーターの速度制御
	duty := dutyMin
	direction := 1.0
	for {
		// デューティ比の設定
		if err := l.SetDuty(duty); err != nil {
			fmt.Printf("failed to set duty: %v\n", err)
			return
		}

		// テレメトリーデータの読み取り
		if telco.WaitForEdge(-1) {
			// エッジが検出されたら、速度を計算
			start := time.Now()
			if telco.WaitForEdge(time.Second) {
				elapsed := time.Since(start)
				rpm := 1.0 / elapsed.Seconds() * 60
				fmt.Printf("Motor speed: %.2f RPM\n", rpm)
			}
		}

		// デューティ比の更新
		duty += 0.01 * direction
		if duty > dutyMax || duty < dutyMin {
			direction *= -1
		}

		time.Sleep(100 * time.Millisecond)
	}
}
