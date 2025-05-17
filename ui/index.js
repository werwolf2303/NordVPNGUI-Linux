const IPCS = {
    GET_ACCOUNT: "get_account",
    GET_LOGIN_URL: "get_login_url",
    GET_COUNTRIES: "get_countries",
    CONNECT: "connect",
    DISCONNECT: "disconnect",
    RECONNECT: "reconnect",
    GET_STATUS: "get_status",
    GET_REGISTER_URL: "get_register_url",
}

function loadPageElement(into, from) {
    const xmlhttp = new XMLHttpRequest();
    xmlhttp.open("GET", from, false);
    xmlhttp.send();
    if (xmlhttp.status === 200) {
        into.appendChild(document.createRange().createContextualFragment(xmlhttp.responseText))
    } else {
        into.innerHTML = "Error while loading";
    }
    window[into.getAttribute("data-init")]()
}

if(typeof debug === "undefined") {
    debug = false;
}

if(!debug) {
    console.log = function (message) {}
    console.debug = function (message) {}
    console.info = function (message) {}
}

document.addEventListener("DOMContentLoaded", () => {
    if(!window.emit) {
        emit = async function (name, arguments, callback, onerror) {
            return await new Promise(resolve => {
                if(debug) {
                    console.log("Calling '" + name + "' with ")
                    if (arguments != null) {
                        for (const argument of arguments) {
                            console.log("Argument: '" + argument + "' null? " + (argument == null))
                        }
                    } else {
                        console.log("Arguments: None")
                    }
                }
                success = ipc.emit(name, arguments, (data) => {
                    if(debug) {
                        console.log("Response: " + data);
                    }
                    if(JSON.parse(data)["Error"]) {
                        onerror(JSON.parse(data)["Error"]["error"]);
                        resolve(success);
                        return;
                    }
                    callback(data);
                    resolve(success);
                });
            });
        }
    }
    /*emit(IPCS.GET_ACCOUNT, null, (data) => {*/
        // User is logged in
        const elements = document.querySelectorAll(
            "[data-element]"
        );
        [ ...elements ]
            .sort((a, b) => a.getAttribute("data-order") > b.getAttribute("data-order"))
            .forEach((element, index) => {
                if(element.getAttribute("data-order") === "-1") return;
                loadPageElement(
                    element,
                    element.getAttribute("data-element") + ".html"
                )
            })
    /*}, (error) => {
        document.getElementById("login_view").style.display = "block";
        loadPageElement(document.getElementById("login_view"),
            document.getElementById("login_view").getAttribute("data-element") + ".html");
        loadPageElement(document.getElementById("error_bar_element"),
            document.getElementById("error_bar_element").getAttribute("data-element") + ".html");
    });*/
});