package stream

// Message 는 스트림이 내보내는 데이터 메시지다. 타입 스위치로 구분한다.
//
// 봉투의 구독 확인(I)·하트비트(H)는 내부에서 소화되며 채널로 나오지 않는다.
// 에러(E)는 Stream.Err() 로 흐른다.
type Message interface{ isMessage() }
