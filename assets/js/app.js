var DEF_PASSCODE = (document.cookie.includes("passcode=")) ? document.cookie.split('; ')
    .find(row => row.startsWith('passcode=')).split("=")[1] : undefined

var LAST_SCAN_TIMESTAMP = (document.cookie.includes("lscan=")) ? document.cookie.split('; ')
    .find(row => row.startsWith('lscan=')).split("=")[1] : undefined

var SERVER = (document.cookie.includes("srv=")) ? document.cookie.split('; ')
    .find(row => row.startsWith('srv=')).split("=")[1] : undefined
var META = {}
var SYNCTABLE = {}


function parseScanContents(scannedContent) {
    const cTrimmed = scannedContent.trim()
    if (cTrimmed.length === 0) return {
        type: "error",
        content: undefined,
        errText: "No data found"
    }

    // Return early if the scanned code is a barcode
    if (cTrimmed.length < 20) {
        return (!isNaN(parseInt(cTrimmed))) ? {
            type: "ean-13",
            content: cTrimmed
        } : {
            type: "error",
            content: undefined,
            errText: "Failed to parse the barcode"
        }
    }

    // Check if the scanned code is a URL
    if (!cTrimmed.includes("https://") || !cTrimmed.includes("?d=")) {
        return {
            type: "error",
            content: undefined,
            errText: "Failed to find the URL"
        }
    }

    let encoded_shit = cTrimmed.substr(cTrimmed.lastIndexOf("?d=") + 3)
    let ticket_type = cTrimmed.split("/")[4]
    try {
        encoded_shit = decodeURIComponent(encoded_shit)
        ticket_type = parseInt(ticket_type)
    } catch (e) {
        return {
            type: "error",
            content: undefined,
            errText: "Failed to parse the URL"
        }
    }
    return {
        type: "qr",
        content: encoded_shit,
        ticket_type_id: ticket_type
    }
}

// listenerSetup sets up a listener on ENTER key, commonly used as "submit" on scanners
function listenerSetup() {
    const area = document.getElementById("scanarea")
    area.focus()
    area.addEventListener('keyup', (e) => {
        if (e.key === 'Enter') {
            console.log(parseScanContents(area.value))
            area.value = ""
        }
    });
}

// apiRequest sends a custom request to the server
async function apiRequest(payload, passcode = DEF_PASSCODE) {

    if (!payload.Passcode && !passcode) {
        return {
            error: "No passcode provided"
        }
    }

    if (!SERVER) {
        return {
            error: "No server configured"
        }
    }

    if (!payload.Passcode) {
        payload.Passcode = passcode
    }

    // Prepare request header vars  
    const content_type = (typeof (payload) === 'object') ? "application/json; charset=UTF-8" : "application/x-www-form-urlencoded"


    const response = await fetch(SERVER + '/req', {
        method: 'POST',
        mode: 'cors', // no-cors, *cors, same-origin, // *default, no-cache, reload, force-cache, only-if-cached
        credentials: 'same-origin',
        cache: 'no-cache',
        headers: {
            'Content-Type': content_type
        },
        redirect: 'follow',
        referrerPolicy: 'no-referrer',
        body: (typeof (payload) === 'object') ? JSON.stringify(payload) : payload
    });
    return response;

}

// config sets some global settings
async function config(passcode, serverUrl) {
    SERVER = serverUrl
    document.cookie = `srv=${SERVER}; SameSite=None; Secure`;
    let ret
    await apiRequest({
            "Handle": "/login",
            "Passcode": passcode,
            "Timestamp": parseInt(Date.now()/1000)
        }).then((r) => r.text().then((t) => JSON.parse(t)))
        .then((jso) => {
            META = {}
            META.valid = jso.code_valid
            if (META.valid) {
                document.cookie = `passcode=${passcode}; SameSite=None; Secure`;
                // Map ticket_type_id to public key
                m = new Map()
                for (const entry of jso.ticket_types) {
                    m.set(entry.re_id, entry.re_key)
                }
                META.ticket_types = m
                ret = {
                    success: true,
                    error: undefined
                }
            } else {
                ret = {
                    success: false,
                    error: "Failed to log in using the code"
                }
            }

        })
    return ret
}

function loginRoutine() {
    if (DEF_PASSCODE) {
        document.getElementById("lb").style.display = "block"
        document.getElementById("mc").style.filter = ""

    } else {
        // Display the login form
        const loginForm = document.getElementById("login")
        const btn = document.getElementById("login_btn")
        const passInput = document.getElementById("passInput")
        const srvInput = document.getElementById("srvInput")
        const lErr = document.getElementById("l-err")
        const mc = document.getElementById("mc")

        loginForm.style.display = "block"
        mc.style.filter = "blur(6px)"
        
        // Bind login logic to a button
        btn.onclick = function () {
            let pass = passInput.value
            let srv = srvInput.value
            if (!pass || !srv) {
                lErr.innerHTML = "Error: Fields cannot be blank"
                return
            }

            try {
                config(pass, srv).then(d => {
                    if (!d.success) {
                        lErr.innerHTML = d.error
                        return
                    } else {
                        // login ok
                        loginForm.style.display = "none"
                        mc.style.filter = ""
                        document.getElementById("lb").style.display = "block"

                    }
                })

            } catch (error) {
                lErr.innerHTML = "Error: login info cannot be validated"
                console.log(error)
                return
            }
        }

        // Clear any errors on input focus
        passInput.onfocus = function () {
            lErr.innerHTML = ""
        }
        srvInput.onfocus = function () {
            lErr.innerHTML = ""
        }

    }
}