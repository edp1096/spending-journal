# 메모

## 구조체
* 지불수단 - Account
* 거래내역 - Record

## 엔드포인트
* DB 초기화 - GET /setup/db

* 거래 추가 - POST /record
* 거래 수정 - PUT /record/update
* 거래 삭제 - DELETE /record/delete
* 거래 목록 - GET /record
* 기간내 거래 내역 - GET /record/sum

* 지불수단 추가 - POST /account
* 지불수단 수정 - PUT /account?id=account:1721395333
* 지불수단 삭제 - DELETE /account?id=account:1721395333
* 지불수단 목록 - GET /account


## 카드 이용기간 (신용공여기간)
* https://www.bccard.com/app/card/ContentsLinkActn.do?pgm_id=ind0623


# Todos
* [x] 지불수단 제어
* [x] 거래내역 입력/수정시 지불수단 없으면 추가
* [x] 환경설정 - 일단 화면 띄우기만
* [x] 비번 변경 - 환경설정 내부에
    * [x] 백단
    * [x] 프론트단
* [ ] 분류 제어(=삭제)
* [ ] 기간 검색 - html
* [ ] 백단 합계 교정/수정
* [ ] ~~sumhandler 제거 및 기간검색 통합~~
* [ ] 통화 변환 - 보류 ~~오늘 날짜 기준으로만, api 시도~~


# 아이콘
* 통장 - https://www.flaticon.com/kr/free-icon/passbook_9235918
* 입출금 - https://www.flaticon.com/kr/free-icon/cash-flow_9235778
* 설정 - https://www.flaticon.com/kr/free-icon/gears_4115432
