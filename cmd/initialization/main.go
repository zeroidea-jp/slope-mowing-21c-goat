package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	rmd "your_package_path/rmd_canbus_v3" // rmd_canbus_v3 パッケージへのパスを適切に設定してください

	"go.einride.tech/can"
	"go.einride.tech/can/pkg/socketcan"
)

const (
	interfaceName     = "can0" // CAN インターフェース名
	initRotations     = 2      // 初期巻き取り回転数
	finalRotations    = 14     // 最終巻き取り回転数
	weakTorqueCurrent = 0.1    // 弱いトルク制御用の電流値 (A)
)

var alignVec = [4]int8{1, -1, -1, 1}

func main() {
	ctx := context.Background()

	// CAN 接続の設定
	conn, err := socketcan.DialContext(ctx, "can", interfaceName)
	if err != nil {
		log.Fatalf("Failed to connect to %s: %v", interfaceName, err)
	}
	defer conn.Close()

	tx := socketcan.NewTransmitter(conn)
	rx := socketcan.NewReceiver(conn)

	// モーターID の入力
	motorID := getMotorIDFromUser()

	// モーターの動作確認
	confirmMotorMovement(ctx, tx, rx, rmd.MotorID(motorID))

	// ユーザーにワイヤー固定の確認を求める
	confirmWireAttachment()

	// 初期巻き取り
	err = performInitialWinding(ctx, tx, rx, rmd.MotorID(motorID))
	if err != nil {
		log.Fatalf("Initial winding failed: %v", err)
	}

	// モーター座標の設定
	err = setMotorPosition(ctx, tx, rx, rmd.MotorID(motorID))
	if err != nil {
		log.Fatalf("Failed to set motor position: %v", err)
	}

	// 最終巻き取り
	err = performFinalWinding(ctx, tx, rx, rmd.MotorID(motorID))
	if err != nil {
		log.Fatalf("Final winding failed: %v", err)
	}

	fmt.Println("Initialization sequence completed successfully.")
}

func getMotorIDFromUser() int {
	var motorID int
	for {
		fmt.Print("Enter target motor ID (1-4): ")
		_, err := fmt.Scanf("%d", &motorID)
		if err == nil && motorID >= 1 && motorID <= 4 {
			break
		}
		fmt.Println("Invalid input. Please enter a number between 1 and 4.")
	}
	return motorID
}

func confirmMotorMovement(ctx context.Context, tx *socketcan.Transmitter, rx *socketcan.Receiver, motorID rmd.MotorID) {
	fmt.Println("Confirming motor movement...")

	// +5度の回転
	err := rmd.SendAbsolutePositionControlCommand(ctx, tx, rx, motorID, rmd.ShaftPositionInRotation(5.0/360.0), rmd.ShaftSpeedInRotPerSec(5.0/360.0))
	if err != nil {
		log.Fatalf("Failed to rotate +5 degrees: %v", err)
	}
	time.Sleep(1 * time.Second)

	// -5度の回転
	err = rmd.SendAbsolutePositionControlCommand(ctx, tx, rx, motorID, rmd.ShaftPositionInRotation(-5.0/360.0), rmd.ShaftSpeedInRotPerSec(5.0/360.0))
	if err != nil {
		log.Fatalf("Failed to rotate -5 degrees: %v", err)
	}
	time.Sleep(1 * time.Second)

	// 0度に戻す
	err = rmd.SendAbsolutePositionControlCommand(ctx, tx, rx, motorID, rmd.ShaftPositionInRotation(0), rmd.ShaftSpeedInRotPerSec(5.0/360.0))
	if err != nil {
		log.Fatalf("Failed to return to 0 degrees: %v", err)
	}
	time.Sleep(1 * time.Second)

	fmt.Println("Motor movement confirmation completed.")
}

func confirmWireAttachment() {
	fmt.Print("Is the wire securely attached to the motor? (y/n): ")
	var response string
	fmt.Scanln(&response)
	if response != "y" && response != "Y" {
		fmt.Println("Please attach the wire and restart the program.")
		os.Exit(1)
	}
}

func performInitialWinding(ctx context.Context, tx *socketcan.Transmitter, rx *socketcan.Receiver, motorID rmd.MotorID) error {
	fmt.Printf("Performing initial winding (%d rotations)...\n", initRotations)
	targetPosition := rmd.ShaftPositionInRotation(float32(initRotations) * float32(alignVec[motorID-1]))
	err := rmd.SendAbsolutePositionControlCommand(ctx, tx, rx, motorID, targetPosition, rmd.ShaftSpeedInRotPerSec(0.5))
	if err != nil {
		return fmt.Errorf("failed to perform initial winding: %v", err)
	}
	err = rmd.WaitForPositionReached(ctx, tx, rx, motorID, targetPosition)
	if err != nil {
		return fmt.Errorf("failed to wait for initial winding completion: %v", err)
	}
	return nil
}

func setMotorPosition(ctx context.Context, tx *socketcan.Transmitter, rx *socketcan.Receiver, motorID rmd.MotorID) error {
	fmt.Println("Setting motor position to 0...")

	// プロトコルに従って、現在の多回転位置をROMにモーターのゼロとして書き込む
	sendingData := can.Data{
		0x64, // Write current multi-turn position to ROM as motor zero command
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
	}

	err := rmd.SendCommand(ctx, tx, motorID, sendingData, rmd.FLAG_SINGLE_MOTOR)
	if err != nil {
		return fmt.Errorf("failed to set motor position: %v", err)
	}

	// モーターの再起動（プロトコルで必要とされている場合）
	fmt.Println("Restarting motor...")
	time.Sleep(2 * time.Second) // 再起動のための待機時間

	// システムリセットコマンドの送信
	resetData := can.Data{
		0x76, // System reset command
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
	}

	err = rmd.SendCommand(ctx, tx, motorID, resetData, rmd.FLAG_SINGLE_MOTOR)
	if err != nil {
		return fmt.Errorf("failed to reset motor: %v", err)
	}

	time.Sleep(5 * time.Second) // リセット後の待機時間

	return nil
}

func performFinalWinding(ctx context.Context, tx *socketcan.Transmitter, rx *socketcan.Receiver, motorID rmd.MotorID) error {
	fmt.Printf("Performing final winding (%d rotations)...\n", finalRotations)

	for i := 0; i < finalRotations; i++ {
		targetPosition := rmd.ShaftPositionInRotation(float32(i+1) * float32(alignVec[motorID-1]))
		err := rmd.SendTorqueControlCommand(ctx, tx, rx, motorID, rmd.CurrentInA(weakTorqueCurrent))
		if err != nil {
			return fmt.Errorf("failed to set weak torque control: %v", err)
		}

		err = rmd.SendAbsolutePositionControlCommand(ctx, tx, rx, motorID, targetPosition, rmd.ShaftSpeedInRotPerSec(0.2))
		if err != nil {
			return fmt.Errorf("failed to perform final winding rotation %d: %v", i+1, err)
		}

		err = rmd.WaitForPositionReached(ctx, tx, rx, motorID, targetPosition)
		if err != nil {
			return fmt.Errorf("failed to wait for final winding rotation %d completion: %v", i+1, err)
		}
	}

	return nil
}
