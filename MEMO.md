# 메모

## 구조체
* 지불수단 - Method
* 거래내역 - Record

## 엔드포인트
* DB 초기화 - GET /setup/db
* 거래 추가 - POST /record
* 거래 검색 - GET /record/search
* 기간내 거래 내역 - GET /record/sum
* 거래 삭제 - DELETE /record/delete
* 거래 수정 - PUT /record/update

* 지불수단 추가
* 지불수단 목록


# Todos
* [ ] 지불수단 제어
* [ ] 거래내역 입력/수정시 지불수단 없으면 추가
* [ ] 통화 변환 - 오늘 날짜 기준으로만, api 시도
