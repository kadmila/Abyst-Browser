let cross = document.createElement("obj", {src:"http://127.0.0.1:80/static/x.obj"});
cross.appendChild(document.createElement("pbrm", {albedo:"http://127.0.0.1:80/static/shard.png"}))

let point = document.createElement("obj", {src:"http://127.0.0.1:80/static/point.obj"});
point.appendChild(document.createElement("pbrm", {albedo:"http://127.0.0.1:80/static/shard.png"}))

const indicator_container = document.getElementById("indicator_container");
function setIndicator(status) {
    const children = indicator_container.children;
    if (children.length > 0) {
        indicator_container.children[0].remove();
    }
    switch (status) {
        case -1:
            indicator_container.appendChild(cross);
            break;
        case 0:
            indicator_container.appendChild(point);
            break;
    }
}

async function register() {
    const response = await fetch("http://127.0.0.1:80/api/register", {
        method: "POST",
        body: host.aurl + "," + host.idCert + "," + host.hsKeyCert,
    });    
    if (response.status !== 200) {
        console.log("failed to register: " 
            + (await response.text()).trim());
        setIndicator(-1);
        return;
    }
    setIndicator(0);

    try {
        while(true) {
            const response = await fetch("http://127.0.0.1:80/api/wait?id=" + host.id);
            if (response.status === 408) { //timeout, no one requested to join.
                continue; //waiting
            } else if (response.status === 200) {
                // received randezvous request
                const body = await response.text();
                const bodyParts = body.split(",", 3);
                if (bodyParts.length != 3) {
                    console.log("failed to parse response: (" + len(bodyParts) + ")");
                    continue;
                }
                console.log(bodyParts[0] + " wants to connect me.")

                host.register(bodyParts[1], bodyParts[2]);
                console.log("(main.aml)registered peer " + bodyParts[0]);

                await sleep(1000);

                console.log("(main.aml)connecting peer " + bodyParts[0]);
                host.connect(bodyParts[0]);
            } else {
                console.log("failed to wait for event: " 
                    + (await response.text()));
                setIndicator(-1);
                return;
            }
        }
    } catch (e) {
        console.log(e.message);
        setIndicator(-1);
    }
}
register();