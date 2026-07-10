package license

// {
//     "code": 0,
//     "error": "",
//     "response": {
//         "key": "TXVB-9ATL-U6YT-FETJ",
//         "data": {
//             "machine_id": "da8a64f92071eda0a394d4b7fbea6fd7cd685b0eb3a31cbb9ee85b869561ae3c",
//             "max_devices": 5,
//             "kind": "normal",
//             "expires_at": 1815057325
//         },
//         "jwt": "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJSaWNobW9iaWxlIiwiYXVkIjoiSGFua2VzdCIsImlhdCI6MTc4MzUyMTMyNSwiZXhwIjoxODE1MDU3MzI1LCJtYWNoaW5lX2lkIjoiZGE4YTY0ZjkyMDcxZWRhMGEzOTRkNGI3ZmJlYTZmZDdjZDY4NWIwZWIzYTMxY2JiOWVlODViODY5NTYxYWUzYyIsImtleSI6IlRYVkItOUFUTC1VNllULUZFVEoiLCJtYXhfZGV2aWNlcyI6NSwia2luZCI6Im5vcm1hbCJ9.HvVYtfWDrzkkkqkG21vgGS3HOvsB5ReL67BxTXfB3plGv8PZcHINV0hJIlnnzUkcgvlClCoUPgJIrMLXU9MG1sAqMhpi-P6llIm0L9fLhM_hnrVJ5Rlp7raNshUusZ9S1OVUayHzWT4WKjOf8mUkyqT_TwpOml5rx3ezW4MUy0H9qvb9DuJoKNkG--Sw4hT0BzYGeXMewgzrEEu2D2oWaqaumkTtwuoD1uX8CSqsKNhlE2LrjACUWzSsXDNB4H9MdR-pzZsfMnWqb60-rz0tXo1oNkArhKx39ns5eCqW0lQznIkaktResC3TeV0HeZ3SHjZcAlHSbWqAZYsTKZIxpSOQSuBqtMVowlCvu_kzq8vUX4fDTtOqcvBwG_x0pw8WIQqCgR_ZRR4FBGeUp0Ymk_1U3Uj52ScTO_vTQ7JQAV4ibyxx2nFbCEjwSAy_yHhbpk-TW6gIsqpBDImY35-Elm1uC2rEJzvtukLorwfM01Vvz9fYSOz0Fo_S-km5b-dYoI6C7lVM9h_4uzWg0Xekjf5w7X1UD3jqO2cUdJWNZbp0vnG2r7waFRj6SiqI1Wvt4IHPhOTlmAgmp9G--uBpdax48IkgRrZwWvYAQeCHYzCUzU3gILxw-EAfyQyKUckksx7f-Z725LQwfLRiOpIAB4OrIAoXqvQrRkKMP8YgrG0"
//     }
// }

type LicenseClaims struct {
	MachineID  string `json:"machine_id"`
	MaxDevices int    `json:"max_devices"`
	Kind       string `json:"kind"`
	ExpiresAt  int    `json:"expires_at"`
}

type LicenseResponseBody struct {
	Key     string `json:"key"`
	JWT     string `json:"jwt"`
	Data    LicenseClaims `json:"data"`
}

type LicenseAPIResponse struct {
	Code     int    `json:"code"`
	Error    string `json:"error"`
	Response LicenseResponseBody `json:"response"`
}

type ApplyRequest struct {
	Kind       string `json:"kind"`
	MachineID  string `json:"machine_id"`
}
