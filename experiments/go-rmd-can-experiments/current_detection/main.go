package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.einride.tech/can"
	"go.einride.tech/can/pkg/socketcan"
)

const MOTOR_ID_1 = MotorID(0x01)
const MOTOR_ID_2 = MotorID(0x02)

const FLAG_SINGLE_MOTOR MotorTargetFlag = 0x140

// const CURRENT_CONTRL_VAL

// const FLAG_MULTI_MOTOR MotorTargetFlag = 0x280

type MotorTargetFlag uint32

// type MotorModelParameters struct {
// }

type MotorID uint8
type Model string

// type MotorModels map[Model]struct {
// }

// type Orders map[MotorID]*OrderData

// type OrderData struct {
// 	canFrame can.Frame
// }

func main() {
	// 制御する全てのモーターの ID を Slice に格納
	// motor_ids := []MotorID{MOTOR_ID_1, MOTOR_ID_2}

	// Context を作成
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	// // シグナルを受信するチャンネルを作成
	// sigCh := make(chan os.Signal, 1)
	// signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// go func() {
	// 	<-sigCh
	// 	cancel()
	// }()

	// CAN 通信のインターフェースを開く
	pConn, err := socketcan.DialContext(ctx, "can", "can0")
	if err != nil {
		log.Println("Error when Dialialing with Context.")
		log.Fatal(err)
	}
	defer pConn.Close()

	// CAN 用の io.Writer の作成
	pTx := socketcan.NewTransmitter(pConn)
	defer pTx.Close()

	// CAN 用の io.Receiver にの作成
	pRx := socketcan.NewReceiver(pConn)
	defer pRx.Close()

	// target_position_in_rot := shaft_position_in_rotation(0) // WORKKING POINT
	// max_speed_in_rps := shaft_speed_in_rot_per_sec(3)

	// _ = send_absolute_postion_control_command(ctx, pTx, pRx, MOTOR_ID_2, target_position_in_rot, max_speed_in_rps)
	// _ = wait_for_position_reached(ctx, pTx, pRx, MOTOR_ID_2, target_position_in_rot)

	// time.Sleep(2 * time.Second)

	max_current_in_A := current_in_A(0.06)

	_ = send_torque_control_command(ctx, pTx, pRx, MOTOR_ID_2, max_current_in_A)

	_ = dangerous_wait_for_shaft_stops(ctx, pTx, pRx, MOTOR_ID_2)

	// time.Sleep(20 * time.Second)

	// max_current_in_A = current_in_A(0)

	// _ = send_torque_control_command(ctx, pTx, pRx, MOTOR_ID_2, max_current_in_A)

	log.Println("Exiting main()")

	// fmt.Printf("Get message: %v\n", msg)
}

//
//
//

type shaft_position_in_rotation float32
type shaft_speed_in_rot_per_sec float32

type current_in_A float32

const ERROR_TOLERANCE_IN_DEGREE = int16(1)
const PRESSING_TOLERANCE_IN_DEGREE = int16(5)
const PRESSING_DURATION_IN_MILLI_S = time.Millisecond * 700

func send_absolute_postion_control_command(ctx context.Context, pTx *socketcan.Transmitter, pRx *socketcan.Receiver, motor_id MotorID, target_position_in_rot shaft_position_in_rotation, max_speed_in_rps shaft_speed_in_rot_per_sec) error {
	speed_alley := max_speed_in_rps.to_deg_per_sec_alley_2x_uint8()        // 1dps/LSB
	position_allay := target_position_in_rot.to_centi_deg_alley_4x_uint8() // 0.01degree/LSB
	log.Println("target_position_in_rot: ", target_position_in_rot)        // DEBUG
	sendingData := can.Data{
		0xA4,              // Command byte
		0x00,              // NULL
		speed_alley[0],    // Speed limit low byte // 1dps/LSB
		speed_alley[1],    // Speed limit low byte
		position_allay[0], // Position control lowest  byte // 0.01degree/LSB
		position_allay[1], // Position control lower   byte
		position_allay[2], // Position control higher  byte
		position_allay[3], // Position control highest byte
	}

	err := send_command(ctx, pTx, motor_id, sendingData, FLAG_SINGLE_MOTOR)
	if err != nil {
		log.Fatal("failed to send position controle command: %", err)
	}

	err = receive_and_check_reply(pRx, motor_id, sendingData)
	if err != nil {
		log.Fatal("failed to receive proper responce: ", err)
	}

	return nil
}

func (shaft_position_in_rotation shaft_position_in_rotation) to_centi_deg_alley_4x_uint8() [4]uint8 {
	degree := int32(shaft_position_in_rotation * 36000)
	return (int32_to_alley_4x_uint8(degree))
}

func (shaft_position_in_rotation shaft_position_in_rotation) to_deg_alley_2x_uint8() [2]uint8 {
	degree := shaft_position_in_rotation.to_degree()
	return (int16_to_alley_2x_uint8(degree))
}

func (shaft_position_in_rotation shaft_position_in_rotation) to_degree() int16 {
	shaft_position_in_degree := int16(shaft_position_in_rotation * 360)
	return shaft_position_in_degree
}

func (shaft_speed_in_rot_per_sec shaft_speed_in_rot_per_sec) to_deg_per_sec_alley_2x_uint8() [2]uint8 {
	degree_per_sec := int16(shaft_speed_in_rot_per_sec * 360)
	return (int16_to_alley_2x_uint8(degree_per_sec))
}

func send_torque_control_command(ctx context.Context, pTx *socketcan.Transmitter, pRx *socketcan.Receiver, motor_id MotorID, max_current_in_A current_in_A) error {
	current_alley := max_current_in_A.to_centiA_alley_2x_uint8()
	sendingData := can.Data{
		0xA1,             // Command byte
		0x00,             // NULL
		0x00,             // NULL
		0x00,             // NULL
		current_alley[0], // Torque control low  byte // 0.01A/LSB
		current_alley[1], // Torque control high byte
		0x00,             // NULL
		0x00,             // NULL
	}

	err := send_command(ctx, pTx, motor_id, sendingData, FLAG_SINGLE_MOTOR)
	if err != nil {
		log.Fatal("failed to send position controle command: %v", err)
	}

	err = receive_and_check_reply(pRx, motor_id, sendingData)
	if err != nil {
		log.Fatal("failed to receive proper responce: %v", err)
	}

	return nil
}

func (current_in_A current_in_A) to_centiA_alley_2x_uint8() [2]uint8 {
	current_in_centiA := int16(current_in_A * 100)
	return (int16_to_alley_2x_uint8(current_in_centiA))
}

func int16_to_alley_2x_uint8(int16_num int16) [2]uint8 {
	return [2]uint8{
		uint8(int16_num & 0xFF),
		uint8(int16_num >> 8 & 0xFF),
	}
}

func int32_to_alley_4x_uint8(int32_num int32) [4]uint8 {
	return [4]uint8{
		uint8(int32_num & 0xFF),
		uint8(int32_num >> 8 & 0xFF),
		uint8(int32_num >> 16 & 0xFF),
		uint8(int32_num >> 24 & 0xFF),
	}
}

func dangerous_wait_for_shaft_stops(ctx context.Context, pTx *socketcan.Transmitter, pRx *socketcan.Receiver, motor_id MotorID) error {

	sendingData := can.Data{
		0x9C, // Command byte: Read Motor Status 2 Command(current, speed, position)
		0x00, // NULL
		0x00, // NULL
		0x00, // NULL
		0x00, // NULL
		0x00, // NULL
		0x00, // NULL
		0x00, // NULL
	}

	var previous_position_in_degree int16 = 0 // CAUTION!! Possible to cause problem
	var unchanged_duration time.Duration

	timer := time.NewTimer(PRESSING_DURATION_IN_MILLI_S)
	defer timer.Stop()

	for {
		interval := 100 * time.Millisecond
		time.Sleep(interval)

		err := send_command(ctx, pTx, motor_id, sendingData, FLAG_SINGLE_MOTOR)
		if err != nil {
			log.Fatal("failed to send position controle command: ", err)
		}

		if pRx.Receive() { // Should be updated so that no reply can be treated

			respFrame := pRx.Frame()

			if respFrame.ID != (0x240+uint32(motor_id)) || respFrame.Data[0] != sendingData[0] {
				return fmt.Errorf("unexpected response received")
			}

			// モーターの現在位置を取得
			current_position_in_degree := uint8_x2_to_int16(respFrame.Data[6], respFrame.Data[7])

			if current_position_in_degree >= previous_position_in_degree-PRESSING_TOLERANCE_IN_DEGREE &&
				current_position_in_degree <= previous_position_in_degree+PRESSING_TOLERANCE_IN_DEGREE {
				unchanged_duration += interval
			} else {
				unchanged_duration = 0
			}
			previous_position_in_degree = current_position_in_degree

			if unchanged_duration >= PRESSING_DURATION_IN_MILLI_S {
				err = send_torque_control_command(ctx, pTx, pRx, motor_id, current_in_A(0))
				if err != nil {
					return fmt.Errorf("failed to send motor stop (set torque 0) command: %v", err)
				}
				log.Printf("Motor %d stopped due to unchanged speed", motor_id)
				return nil
			}

			// log.Printf("Motor %d current position: %d", motor_id, current_position_in_degree)

			// leave logs
			err = log_motion(respFrame)
			if err != nil {
				log.Println("Error when loging motion: ", err)
			}
		}

	}

}

func wait_for_position_reached(ctx context.Context, pTx *socketcan.Transmitter, pRx *socketcan.Receiver, motor_id MotorID, target_position_in_rot shaft_position_in_rotation) error {

	sendingData := can.Data{
		0x9C, // Command byte: Read Motor Status 2 Command(current, speed, position)
		0x00, // NULL
		0x00, // NULL
		0x00, // NULL
		0x00, // NULL
		0x00, // NULL
		0x00, // NULL
		0x00, // NULL
	}

	target_position_in_degree := target_position_in_rot.to_degree()

	for {
		// interfal duration
		time.Sleep(100 * time.Millisecond)

		// sending command
		err := send_command(ctx, pTx, motor_id, sendingData, FLAG_SINGLE_MOTOR)
		if err != nil {
			log.Fatal("failed to send position controle command: ", err)
		}

		// receiving responce
		if pRx.Receive() { // Should be updated so that no reply can be treated

			respFrame := pRx.Frame()

			if respFrame.ID != (0x240+uint32(motor_id)) || respFrame.Data[0] != sendingData[0] {
				return fmt.Errorf("unexpected response received")
			}

			// モーターの現在位置を取得
			current_position_in_degree := uint8_x2_to_int16(respFrame.Data[6], respFrame.Data[7])

			if current_position_in_degree >= target_position_in_degree-ERROR_TOLERANCE_IN_DEGREE &&
				current_position_in_degree <= target_position_in_degree+ERROR_TOLERANCE_IN_DEGREE {
				log.Printf("Motor %d reached (target position: %.2f in rotation)", motor_id, target_position_in_rot)
				return nil
			}

			// log.Printf("Motor %d current position: %d", motor_id, current_position_in_degree)

			// leave logs
			err = log_motion(respFrame)
			if err != nil {
				log.Println("Error when loging motion: ", err)
			}
		}

	}
}

func uint8_x2_to_int16(byte_low, byte_high uint8) int16 {
	return (int16(byte_low) | (int16(byte_high) << 8))
}

func send_motor_shutdown_command(ctx context.Context, pTx *socketcan.Transmitter, pRx *socketcan.Receiver, motor_id MotorID) error {
	sendingData := can.Data{
		0x80, // Command byte: To shutdown motor
		0x00, // NULL
		0x00, // NULL
		0x00, // NULL
		0x00, // NULL
		0x00, // NULL
		0x00, // NULL
		0x00, // NULL
	}
	err := send_command(ctx, pTx, motor_id, sendingData, FLAG_SINGLE_MOTOR)
	if err != nil {
		log.Fatal("failed to send position controle command: %v", err)
	}

	err = receive_and_check_reply(pRx, motor_id, sendingData)
	if err != nil {
		log.Fatal("failed to receive proper responce: %v", err)
	}

	return nil
}

func log_motion(frame can.Frame) error {

	if frame.Data[0] == uint8(0x9C) {
		// ログファイル名を生成

		filename := fmt.Sprintf("log/can_log_%X_%X.log", frame.ID, frame.Data[0])

		// ログファイルを開く（または作成する）
		file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println("Error when Opening File")
			return err
		}
		defer file.Close()

		// 電流の計算
		torque_current_in_A := float32(int16(frame.Data[2])|(int16(frame.Data[3])<<8)) / 100      // "100" comes from 0.01A/LSB
		shaft_speed_in_rev_p_sec := float32(int16(frame.Data[4])|(int16(frame.Data[5])<<8)) / 360 // 1dps/LSB
		shaft_position_in_rev := float32(int16(frame.Data[6])|(int16(frame.Data[7])<<8)) / 360    // 1dps/LSB
		// フレームの情報をログファイルに書き込む
		log.SetOutput(file)
		log.Printf("ID: %X, Data: %X, Celsius: %d, A: %+.2f, rev/s: %+.3f,  rev: %+.3f\n", frame.ID, frame.Data, frame.Data[1], torque_current_in_A, shaft_speed_in_rev_p_sec, shaft_position_in_rev)
	} else {
		fmt.Printf("ID: %X, Data: %X\n", frame.ID, frame.Data)
	}
	return nil
}

func send_command(ctx context.Context, pTx *socketcan.Transmitter, motor_id MotorID, data can.Data, flag MotorTargetFlag) error {
	frame := can.Frame{
		ID:         uint32(motor_id) | uint32(flag),
		Length:     8,
		Data:       data,
		IsRemote:   false,
		IsExtended: false,
	}

	log.Print("Sending frame: ")
	log.Println(frame)

	err := pTx.TransmitFrame(ctx, frame)
	if err != nil {
		log.Println("Error when transmitting Frame (1)")
	}
	return err
}

func receive_and_check_reply(pRx *socketcan.Receiver, motor_id MotorID, sendingData can.Data) error {

	if pRx.Receive() { // Should be updated so that no reply can be treated
		respFrame := pRx.Frame()
		log.Print("Replied frame: ")
		log.Println(respFrame)
		if respFrame.ID != (0x240+uint32(motor_id)) || respFrame.Data[0] != sendingData[0] {
			return fmt.Errorf("unexpected response received")
		}
	}
	return nil
}

// func (deci_A current_in_deci_A_int16) to_can_data_alley() [2]uint8 {
// 	return (int16_to_uint8_alley(int16(deci_A)))
// 	// return integer_to_alley_uint8_x_2(deci_A)
// }

// func (deci_deg angle_in_deci_degree_int16) to_can_data_alley() [2]uint8 {
// 	return (int16_to_uint8_alley(int16(deci_deg)))
// 	// return integer_to_alley_uint8_x_2(deci_deg)
// }

// func (deci_deg angle_in_deci_degree_int32) to_can_data_alley() [4]uint8 {
// 	return (int32_to_uint8_alley(int32(deci_deg)))
// 	// return integer_to_alley_uint8_x_4(deci_deg)
// }

// func (deg_per_sec shaft_angle_speed_in_deg_per_sec_int16) to_can_data_alley() [2]uint8 {
// 	return (int16_to_uint8_alley(int16(deg_per_sec)))
// 	// return integer_to_alley_uint8_x_2(deg_per_sec)
// }
