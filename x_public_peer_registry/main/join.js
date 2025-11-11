console.log("hello join.js!");

const point_obj = document.getElementById("point");
function setFailPoint() {
    point_obj.src = "http://127.0.0.1:80/static/x.obj";
}

async function requestRandom() {
    try{
        const random_resp = await fetch(`http://127.0.0.1:80/api/random?excl=${host.id}`);
        if (random_resp.status !== 200) {
            console.log(`failed to fetch random join target: ${random_resp.statusText}:${(await random_resp.text()).trim()}`);
            setFailPoint();
            return;
        }

        const target = await random_resp.text();
        console.log("(join.js)target peer: " + target);

        const join_resp = await fetch(`http://127.0.0.1:80/api/request?id=${host.id}&targ=${target}`);
        if (join_resp.status !== 200) {
            console.log(`failed to fetch join request: ${join_resp.statusText}:${(await join_resp.text()).trim()}`);
            setFailPoint();
            return;
        }

        const join_data_raw = await join_resp.text()
        //console.log("(join.js)data: " + join_data_raw);

        const join_data = join_data_raw.split(",");
        host.register(join_data[1], join_data[2]);
        console.log("(join.js)registered peer " + join_data[0]);

        await sleep(1000);

        //this should be unnecessary
        console.log("(join.js)connecting peer " + join_data[0]);
        host.connect(join_data[0]);

        await sleep(1000);
        
        console.log("(join.js)moving world");
        host.move_world(join_data[0]);
    } catch (e) {
        console.log(e);
    }
}

requestRandom();