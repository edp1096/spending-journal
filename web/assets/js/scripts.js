const url = window.location

const protocol = url.protocol
const hostname = url.hostname
const port = url.port || (protocol === 'https:' ? '443' : '80')

// const addr = "http://localhost:8080"
const addr = `${protocol}//${hostname}:${port}`

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
    "hybrid": "복합",
}

const numericPointStep = { "USD": "0.01", "KRW": "1" }

let exchangeRate = 1300

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

function convertDateMD(dateString) {
    if (!dateString) return ''
    const [year, month, day] = dateString.split('-')

    return `${month}-${day}`
}


const allData = () => {
    const initializer = {
        appReady: false,
        lightmode: JSON.parse(localStorage.getItem("lightmode")),

        chart: null,
        chartCredit: null,
        showHomeScreen: false,
        showAccountList: false,
        showCategoryList: false,
        showRecordList: false,

        summaryDateInterval: parseInt(localStorage.getItem("summary-date-interval") || "7"),
        summaryDateFrom: new Date(new Date().setDate(new Date().getDate() - 7)).toLocaleDateString('en-CA'),
        summaryDateTo: new Date().toLocaleDateString('en-CA'),

        accounts: [],
        accountData: { open: false, account: {} },
        categories: [],
        categoryData: { open: false, category: {} },
        recordsResponse: {},
        recordData: {
            open: false,
            step: numericPointStep,
            accountName: "",
            record: {}
        },
        preferenceData: { open: false, preferences: { "old-password": "", "new-password": "" } },

        clearListViewSelection() {
            const container = document.querySelector(".header-button-container").children
            for (const c of container) { c.classList.remove("contrast") }
            this.showHomeScreen = false
            this.showAccountList = false
            this.showCategoryList = false
            this.showRecordList = false
        },

        /* Home screen */
        setupChart() {
            if (this.chart) { this.chart.destroy() }

            const total = parseFloat(this.recordsResponse["sum-pay"]) + parseFloat(this.recordsResponse["sum-credit-pay"])
            const labels = [
                ...new Set([
                    ...Object.keys(this.recordsResponse["stats"]).map(
                        key => this.recordsResponse["stats"][key].category
                    ),
                    ...Object.keys(this.recordsResponse["stats-credit"]).map(
                        key => this.recordsResponse["stats-credit"][key].category
                    )
                ])
            ]
            const values = labels.map(label => {
                const amountDirect = Object.values(this.recordsResponse["stats"]).find(
                    item => item.category === label
                )?.amount || 0
                const amountCredit = Object.values(this.recordsResponse["stats-credit"]).find(
                    item => item.category === label
                )?.amount || 0

                return ((amountDirect + amountCredit) / total * 100).toFixed(0)
            })

            const chartData = {
                labels: labels,
                datasets: [{ label: '지출 비중', data: values, borderWidth: 1 }],
            }

            Chart.overrides.pie.plugins.legend.position = "right"
            const chartCTX = document.querySelector("#home-chart")
            const chartCreditCTX = document.querySelector("#home-chart-credit")
            if (labels.length > 0) {
                this.chart = new Chart(chartCTX, {
                    type: "pie",
                    data: chartData,
                    options: { plugins: { tooltip: { callbacks: { label: (ctx) => { return `${ctx.label}: ${ctx.parsed}%` } } } } }
                })
            }
        },
        async showHome() {
            this.clearListViewSelection()
            this.showHomeScreen = true
            await this.$nextTick()
            this.setupChart()

            return
        },

        /* Account control */
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
        openInputAccount(index = null) {
            this.accountData.account = {}

            this.accountData.account["account-name"] = ""
            this.accountData.account["pay-type"] = "direct"
            this.accountData.account["repay-day"] = ""
            this.accountData.account.description = ""

            if (index >= 0) {
                this.accountData.account["account-id"] = this.accounts[index].id
                this.accountData.account["account-name"] = this.accounts[index]["account-name"]
                this.accountData.account["pay-type"] = this.accounts[index]["pay-type"]
                this.accountData.account["repay-day"] = this.accounts[index]["repay-day"]
                this.accountData.account["use-day-from"] = this.accounts[index]["use-day-from"]
                this.accountData.account["use-day-to"] = this.accounts[index]["use-day-to"]
                this.accountData.account.description = this.accounts[index].description
            }

            this.accountData.open = true
        },
        async requestSetAccount() {
            let requestMethod = "POST"
            let params = ""
            if (this.accountData.account["account-id"]) {
                requestMethod = "PUT"
                params = `?id=${this.accountData.account["account-id"]}`
            }

            const uri = `${addr}/account${params}`
            const r = await fetch(uri, {
                method: requestMethod,
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

            alert("Fail to set account")
            return false
        },
        async setAccount(event) {
            if (!checkFormValidation(this.$refs.accountForm, event)) { return false }

            if (this.isAccountNameDuplicate()) {
                alert("같은 이름의 계정이 있습니다.")
                return false
            }

            await this.requestSetAccount()
        },
        async deleteAccount(index) {
            if (!this.accounts[index].id) {
                alert("Wrong action")
                return false
            }

            const uri = `${addr}/account?id=${this.accounts[index].id}`
            const r = await fetch(uri, { method: "DELETE" })
            if (r.ok) {
                const response = await r.json()

                if (response.status == "success") {
                    await this.getAccounts()
                    this.accountData.open = false
                    return
                }
            }

            alert("Fail to delete account")
            return false
        },

        /* Category control */
        async getCategories() {
            const uri = `${addr}/category`
            const r = await fetch(uri)
            if (r.ok) {
                this.categories = await r.json()
                return true
            }

            return false
        },
        async showCategories() {
            if (await this.getCategories()) {
                this.clearListViewSelection()
                this.showCategoryList = true
                this.$event.target.classList.add('contrast')
            }
        },
        openInputCategory(index = null) {
            this.categoryData.category = {}

            this.categoryData.category["category-name"] = ""

            if (index >= 0) {
                this.categoryData.category["category-id"] = this.categories[index].id
                this.categoryData.category["category-name"] = this.categories[index]["category-name"]
            }

            this.categoryData.open = true
        },
        async requestSetCategory() {
            let requestMethod = "POST"
            let params = ""
            if (this.categoryData.category["category-id"]) {
                requestMethod = "PUT"
                params = `?id=${this.categoryData.category["category-id"]}`
            }

            const uri = `${addr}/category${params}`
            const r = await fetch(uri, {
                method: requestMethod,
                headers: { "content-Type": "application/json" },
                body: JSON.stringify(this.categoryData.category)
            })
            if (r.ok) {
                const response = await r.json()

                if (response.status == "success") {
                    await this.getCategories()
                    this.categoryData.open = false
                    return
                }
            }

            alert("Fail to set category")
            return false
        },
        async setCategory(event) {
            if (!checkFormValidation(this.$refs.categoryForm, event)) { return false }

            if (this.isCategoryNameDuplicate()) {
                alert("같은 이름의 분류가 있습니다.")
                return false
            }

            await this.requestSetCategory()
        },
        async deleteCategory(index) {
            if (!this.categories[index].id) {
                alert("Wrong action")
                return false
            }

            const uri = `${addr}/category?id=${this.categories[index].id}`
            const r = await fetch(uri, { method: "DELETE" })
            if (r.ok) {
                const response = await r.json()

                if (response.status == "success") {
                    await this.getCategories()
                    this.categoryData.open = false
                    return
                }
            }

            alert("Fail to delete category")
            return false
        },

        /* Record control */
        async getRecords() {
            if (!this.appReady) { return false }
            if (isNaN(new Date(this.summaryDateFrom).getTime()) || isNaN(new Date(this.summaryDateTo).getTime())) {
                return false
            }

            // const uri = `${addr}/record?q=record:&pageSize=1000`
            const uri = `${addr}/record?q=record:&from=${this.summaryDateFrom}&to=${this.summaryDateTo}`
            const r = await fetch(uri)
            if (r.ok) {
                this.recordsResponse = await r.json()
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
        openInputRecord(index = null) {
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

            if (index >= 0) {
                this.recordData.record["transaction-type"] = this.recordsResponse.records[index]["transaction-type"]
                this.recordData.record["pay-type"] = this.recordsResponse.records[index]["pay-type"]

                for (a of this.accounts) {
                    if (a.id == this.recordsResponse.records[index]["account-id"]) {
                        this.recordData.accountName = a["account-name"]
                        this.recordData.record["account-id"] = a.id
                        break
                    }
                }

                this.recordData.record.id = this.recordsResponse.records[index].id
                this.recordData.record.category = this.recordsResponse.records[index].category
                this.recordData.record.currency = this.recordsResponse.records[index].currency
                this.recordData.record.amount = this.recordsResponse.records[index].amount
                this.recordData.record.date = this.recordsResponse.records[index].date
                this.recordData.record.time = this.recordsResponse.records[index].time
                this.recordData.record.description = this.recordsResponse.records[index].description
            }

            this.recordData.open = true
        },
        async requestSetRecord() {
            let requestMethod = "POST"
            let params = ""
            if (this.recordData.record.id) {
                requestMethod = "PUT"
                params = `?id=${this.recordData.record.id}`
            }

            const uri = `${addr}/record${params}`
            const r = await fetch(uri, {
                method: requestMethod,
                headers: { "Content-Type": "application/json" },
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

            let isCategoryExist = false
            for (c of this.categories) {
                if (c["category-name"] == this.recordData.record["category"]) {
                    isCategoryExist = true
                    break
                }
            }

            if (!isCategoryExist) {
                this.categoryData.category = {
                    "category-name": this.recordData.record["category"]
                }
                await this.requestSetCategory()
            }

            await this.requestSetRecord()
        },
        async deleteRecord(index) {
            if (!this.recordsResponse.records[index].id) {
                alert("Wrong action")
                return false
            }

            const uri = `${addr}/record?id=${this.recordsResponse.records[index].id}`
            const r = await fetch(uri, { method: "DELETE" })
            if (r.ok) {
                const response = await r.json()

                if (response.status == "success") {
                    await this.getRecords()
                    this.recordData.open = false
                    return
                }
            }

            alert("Fail to delete record")
            return false
        },
        async changeDateRange() {
            await this.getRecords()
            if (this.showHomeScreen) { this.showHome() }
        },
        isAccountNameDuplicate() {
            let result = false

            if (this.accountData.account["account-name"] == "") {
                return result
            }

            for (const a of this.accounts) {
                if (a["account-name"] == this.accountData.account["account-name"]) {
                    if (a.id == this.accountData.account["account-id"]) {
                        result = false // 같은 row는 수정이므로 중복체크 안하고 패스
                        break
                    }

                    result = true
                }
            }

            return result
        },
        isCategoryNameDuplicate() {
            let result = false

            if (this.categoryData.category["category-name"] == "") {
                return result
            }

            for (const c of this.categories) {
                if (c["category-name"] == this.categoryData.category["category-name"]) {
                    if (c.id == this.categoryData.category["category-id"]) {
                        result = false // 같은 row는 수정이므로 중복체크 안하고 패스
                        break
                    }

                    result = true
                }
            }

            return result
        },
        async setRecordPayType() {
            for (const a of this.accounts) {
                if (a["account-name"] == this.accountName) {
                    this.recordData.record["pay-type"] = a["pay-type"]
                    if (a["pay-type"] == "hybrid") {
                        this.recordData.record["pay-type"] = "direct"
                    }
                    break
                }
            }
        },
        async changePassword(event) {
            if (!checkFormValidation(this.$refs.preferenceForm, event)) { return false }

            const passwordOLD = this.preferenceData.preferences["old-password"]
            const passwordNEW = this.preferenceData.preferences["new-password"]
            const params = `?old-password=${passwordOLD}&new-password=${passwordNEW}`

            const uri = `${addr}/setup/db/password${params}`
            const r = await fetch(uri)
            if (r.ok) {
                const response = await r.json()

                if (response.status == "success") {
                    alert("Password is changed")
                    this.preferenceData.open = false
                    openPasswordGate()
                    return
                }
            }

            alert("Fail to change password")
            return false
        },
        openPreference() {
            this.preferenceData.preferences = {
                "old-password": "",
                "new-password": ""
            }

            this.preferenceData.open = true
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
            body.appReady = true
            body.getAccounts()
            body.getCategories()
            body.getRecords()
            body.showHome()

            return
        }
    }

    alert("Wrong password")
    return false
}

document.addEventListener('alpine:init', () => { })



