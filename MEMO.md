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


# Todos
* [x] 지불수단 제어
* [x] 거래내역 입력/수정시 지불수단 없으면 추가
* [ ] 비번 변경
* [ ] 환경설정
* [ ] 기간 검색 - html
* [ ] 합계 교정
* [ ] sumhandler 제거 및 기간검색 통합
* [ ] 통화 변환 - 오늘 날짜 기준으로만, api 시도