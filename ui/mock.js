if (typeof ipc == "undefined") {
    ipc = class {
        emit = function () {
        };
    }
}

debug = true;

emit = async function (name, argument, callback, onerror) {
    return await new Promise(resolve => {
        setTimeout(() => {
            resolve(ipc.emit(name, argument, (data) => {
                if(JSON.parse(data)["Error"]) {
                    onerror(JSON.parse(data)["Error"]["error"]);
                    return;
                }
                callback(data);
            }));
        }, /*1500*/ 0);
    });
}

ipc.emit = function (name, arguments, callback) {
    console.log("Calling '" + name + "' with ")
    if(arguments != null) {
        for (const argument of arguments) {
            console.log("Argument: '" + argument + "' null? " + (argument == null))
        }
    }else{
        console.log("Arguments: None")
    }
    switch (name) {
        case IPCS.GET_ACCOUNT: {
            /*callback(JSON.stringify({
                "type": 1000,
                "email": "matthiasbraun@gmail.com",
                "expires_at": "2026-05-24 16:12:22",
                "dedicated_ip_status": 3020,
                "mfa_status": 1
            }))*/
            callback(JSON.stringify({
                Error: {error: "rpc error: code = Unknown desc = you are not logged in"}
            }));
            break;
        }
        case IPCS.GET_LOGIN_URL: {
            callback("{\"Login\": { \"url\": \"https://api.nordvpn.com/v1/users/oauth/login-redirect?attempt=a03772fb-d639-4a87-a4d2-2da91ea394ae\" }}")
            break;
        }
        case IPCS.GET_COUNTRIES: {
            const xmlhttp = new XMLHttpRequest();
            xmlhttp.open("GET", "sample.json", false);
            xmlhttp.send();
            if (xmlhttp.status === 200) {
                callback(xmlhttp.responseText)
            }
            break;
        }
        case IPCS.CONNECT: {
            name = arguments[0];
            if(arguments[1]) {
                name = arguments[1];
            }
            callback(JSON.stringify({
                ConnectionInfo: [ name + " #1089", "de1089.nordvpn.com", ""]
            }))
            break;
        }
        case IPCS.DISCONNECT: {
            callback("true");
            break;
        }
        case IPCS.RECONNECT: {
            callback(JSON.stringify({
                ConnectionInfo: ["Germany #69", "de1089.nordvpn.com", ""]
            }))
            break;
        }
        case IPCS.GET_STATUS: {
            //callback('{"technology":2,"protocol":1,"ip":"212.23.215.13","hostname":"de1152.nordvpn.com","country":"Germany","city":"Frankfurt","download":11356,"upload":6244,"uptime":2603361746,"name":"Germany #1152","parameters":{"source":1}}')
            callback("{\"uptime\": -1}")
            break;
        }
        case IPCS.GET_REGISTER_URL: {
            callback("{\"Login\": { \"url\": \"https://napps-2.com/v1/users/oauth/login-redirect?attempt=670f9a38-46f6-4444-9765-112931972277&nord_origin=nordvpn_windows&utm_source=windows&utm_medium=in-app&utm_campaign=desktop-app\" }}")
        }
        default:
            return false;
    }
    return true;
}