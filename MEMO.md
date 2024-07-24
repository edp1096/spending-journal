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
* [x] 백단 합계 교정/수정
* [x] 기간 검색
* [x] 분류 제어
    * [x] 삭제용 목록창, 삭제
    * [x] 기록 입력/수정시 datalist
* [ ] 달러 계산
    * [ ] 통화 변환 - 일단 하드코딩, 보류: ~~오늘 날짜 기준으로만, api 시도~~


# 아이콘
* 통장 - https://www.flaticon.com/kr/free-icon/passbook_9235918
* 입출금 - https://www.flaticon.com/kr/free-icon/cash-flow_9235778
* 설정 - https://www.flaticon.com/kr/free-icon/gears_4115432


## swal
* https://www.codeply.com/p/ysgHFaqFYN


# 홈 인사말 적었던거
```html
<div>
    간단한 지출기입장입니다.
</div>
<div>&nbsp;</div>
<div>
    신용카드 결제 계산을 위한 설정에 손이 덜 가게끔 카드이용일과 정산일 만으로 현재와 과거, 미래의 지출 상황이 어떻게 되는지 알 수 있게 만들었습니다.
</div>
<div>&nbsp;</div>
<div>
    `계정`과 `분류`는 평소에는 건드릴 필요 없이 거래탭에서 입력된 내용으로부터 새로운 항목이 있으면 자동으로 입력됩니다. 신용카드의 경우는 계정탭으로 이동하여 이용기간과 결제일만 수정해주면 됩니다.
</div>
<div>&nbsp;</div>
<div>
    이용내역은 화면 아래쪽 기간섹션의 날짜 범위를 지정하여 입출금 합계를 확인할 수 있습니다.
</div>
<div>&nbsp;</div>
<div>
    본 화면은 나중에는 이 인사말을 없애고 차트를 넣을겁니다.
</div>
```