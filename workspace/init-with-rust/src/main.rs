use std::io;
use std::process;
use socketcan::{CanFrame, CanSocket, Socket, EmbeddedFrame, StandardId};

fn main() {

    let interface = "can0";

    // CANソケットを開く
    let socket = CanSocket::open(interface)
        .expect(&format!("CANソケット {} を開けませんでした", interface));

    // println!("debug 0-0");
    // 現在のモーターIDを読み取るコマンドを送信
    let id_for_settings = StandardId::new(0x123).expect("Failed to create StandardId");
    let frame = CanFrame::new(id_for_settings, &[0x79, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00])
        .expect("フレーム生成に失敗しました");
    socket.write_frame(&frame).expect("コマンド送信に失敗しました");


    println!("debug 0-1");
    // 応答フレームを受信
    let frame = socket.read_frame().expect("応答受信に失敗しました");

    println!("debug 0-2");
    // 現在のモーターIDを表示
    let data = frame.data();
    if data[0] == 0x79 && data[2] == 0x01 {
        let current_id = ((data[7] as u16) << 8) | (data[6] as u16);
        println!("現在のモーターIDは {} です", current_id);

        // ユーザーにIDの設定を促す
        println!("新しいモーターIDを設定しますか? (y/n)");
        let mut input = String::new();
        io::stdin().read_line(&mut input).expect("入力の読み取りに失敗しました");
        if input.trim().to_lowercase() != "y" {
            println!("モーターIDの設定をスキップします");
            return;
        }
    } else {
        eprintln!("モーターIDの読み取りに失敗しました");
        process::exit(1);
    }

    println!("debug 0-3");
    // ユーザーに新しいIDを入力してもらう
    println!("新しいモーターIDを入力してください (1-32):");
    let mut input = String::new();
    io::stdin().read_line(&mut input).expect("入力の読み取りに失敗しました");
    let motor_id = match input.trim().parse::<u8>() {
        Ok(id) if (1..=32).contains(&id) => id,
        _ => {
            eprintln!("モーターIDは1から32の範囲で指定してください");
            process::exit(1);
        }
    };

    // モーターIDを設定するコマンドを送信
    let data = [0x79, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, motor_id];
    let frame = CanFrame::new(id_for_settings, &data)
        .expect("フレーム生成に失敗しました");
    socket.write_frame(&frame).expect("コマンド送信に失敗しました");

    // 応答フレームを受信
    let frame = socket.read_frame().expect("応答受信に失敗しました");

    // 応答をチェック
    let data = frame.data();
    if data[0] == 0x79 && data[2] == 0x00 && &data[4..8] == &[0x00, 0x00, 0x00, motor_id] {
        println!("モーターID {} の設定に成功しました", motor_id);
    } else {
        println!("モーターID {} の設定に失敗しました", motor_id);
    }
}