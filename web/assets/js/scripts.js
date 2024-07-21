const addr = "http://localhost:8080"

const currencyCode = {
    "KRW": "₩",
    "USD": "$",
    "JPY": "¥",
    "EUR": "€",
    "GBP": "£",
    "CNY": "Y",
}

const payType = {
    "direct": "선불",
    "credit": "신용",
}

function checkFormValidation(form, event) {
    event.preventDefault()
    if (event.key == "Enter" && (event.ctrlKey || event.altKey)) {
        return false
    }

    if (!form.checkValidity()) {
        form.reportValidity()
        return false
    }

    return true
}

const allData = () => {
    const initializer = {
        showAccountList: false,
        showRecordList: false,
        accounts: [],
        accountData: {
            open: false,
            account: {}
        },
        records: {},
        recordData: {
            open: false,
            numberStep: { "USD": "0.01", "KRW": "1" },
            accountName: "",
            record: {}
        },
        clearListViewSelection() {
            const container = document.querySelector(".header-button-container").children
            for (const c of container) { c.classList.remove("contrast") }
            this.showAccountList = false
            this.showRecordList = false
        },
        async getAccounts() {
            const uri = `${addr}/account`
            const r = await fetch(uri)
            if (r.ok) {
                this.accounts = await r.json()
                return true
            }

            return false
        },
        async showAccounts() {
            if (await this.getAccounts()) {
                this.clearListViewSelection()
                this.showAccountList = true
                this.$event.target.classList.add('contrast')
            }
        },
        openInputAccount() {
            this.accountData.account = {}

            this.accountData.account["account-name"] = ""
            this.accountData.account["pay-type"] = "direct"
            this.accountData.account["repay-day"] = ""
            this.accountData.account.description = ""

            this.accountData.open = true
        },
        async requestSetAccount() {
            const uri = `${addr}/account`
            const r = await fetch(uri, {
                method: "POST",
                headers: { "content-Type": "application/json" },
                body: JSON.stringify(this.accountData.account)
            })
            if (r.ok) {
                const response = await r.json()

                if (response.status == "success") {
                    await this.getAccounts()
                    this.accountData.open = false
                    return
                }
            }

            alert("Fail to set record")
            return false
        },
        async setAccount(event) {
            if (!checkFormValidation(this.$refs.accountForm, event)) { return false }

            this.accountData.account.amount = parseFloat(this.accountData.account.amount)
            if (this.accountData.account.currency == "KRW") {
                this.accountData.account.amount = parseInt(this.accountData.account.amount)
            }

            await this.requestSetAccount()
        },
        async getRecords() {
            const uri = `${addr}/record?q=record:&pageSize=1000`
            const r = await fetch(uri)
            if (r.ok) {
                this.records = await r.json()
                return true
            }

            return false
        },
        async showRecords() {
            if (await this.getRecords()) {
                this.clearListViewSelection()
                this.showRecordList = true
                this.$event.target.classList.add('contrast')
            }
        },
        openInputRecord() {
            this.recordData.accountName = ""

            this.recordData.record = {}
            this.recordData.record["account-id"] = ""
            this.recordData.record["transaction-type"] = "record_type_pay"
            this.recordData.record["pay-type"] = "direct"
            this.recordData.record.currency = "KRW"
            this.recordData.record.description = ""

            const now = new Date()
            const currentDate = now.toLocaleDateString([], { year: 'numeric', month: '2-digit', day: '2-digit' }).replace(/\.\s/g, '-').replace(/\./g, '')
            const currentTime = now.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false })

            this.recordData.record.date = currentDate
            this.recordData.record.time = currentTime

            this.recordData.open = true
        },
        async requestSetRecord() {
            const uri = `${addr}/record`
            const r = await fetch(uri, {
                method: "POST",
                headers: { "content-Type": "application/json" },
                body: JSON.stringify(this.recordData.record)
            })
            if (r.ok) {
                const response = await r.json()

                if (response.status == "success") {
                    await this.getRecords()
                    this.recordData.open = false
                    return
                }
            }

            alert("Fail to set record")
            return false
        },
        async setRecord(event) {
            if (!checkFormValidation(this.$refs.recordForm, event)) { return false }

            this.recordData.record.amount = parseFloat(this.recordData.record.amount)
            if (this.recordData.record.currency == "KRW") {
                this.recordData.record.amount = parseInt(this.recordData.record.amount)
            }

            for (a of this.accounts) {
                if (a["account-name"] == this.accountName) {
                    this.recordData.record["account-id"] = a.id
                    break
                }
            }

            if (this.recordData.record["account-id"] == "") {
                this.accountData.account = {
                    "account-name": this.accountName,
                    "pay-type": "direct",
                    "repay-day": "",
                    "description": "",
                }
                await this.requestSetAccount()

                for (a of this.accounts) {
                    if (a["account-name"] == this.accountName) {
                        this.recordData.record["account-id"] = a.id
                        break
                    }
                }
            }

            await this.requestSetRecord()
        },
        init() { }
    }

    return initializer
}

function openPasswordGate() {
    const passwordGate = Alpine.$data(document.querySelector("#password-gate-container"))
    passwordGate.open = true

    document.querySelector("#password").value = ""
}

async function enterPassword(event) {
    const passwordGate = Alpine.$data(document.querySelector("#password-gate-container"))
    if (!checkFormValidation(passwordGate.$refs.passwordForm, event)) { return false }

    const password = document.querySelector("#password").value
    const uri = `${addr}/setup/db?password=${password}`

    const r = await fetch(uri)
    if (r.ok) {
        const response = await r.json()
        if (response.status == "success") {
            passwordGate.open = false

            const body = Alpine.$data(document.querySelector("body"))
            body.getAccounts()
            body.getRecords()

            return
        }
    }

    alert("Wrong password")
    return false
}

document.addEventListener('alpine:init', () => { })