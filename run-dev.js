const child_process = require("child_process");
const fs = require("fs");
require("colors");

function stamp() {
  return new Date().toLocaleString();
}

let child;

function run() {
  child = child_process.spawn(
    "go",
    ["run", __dirname + "/src/gothic/main.go"],
    {
      stdio: "inherit",
      cwd: __dirname
    }
  );

  child.on("close", function(code) {
    child.kill("SIGINT");
    run();
  });
}

const dirs = [__dirname + "/src"];

function recursive(base) {
  for (let dir of fs.readdirSync(base)) {
    dir = base + "/" + dir;
    const stat = fs.lstatSync(dir);
    if (stat.isDirectory()) {
      recursive(dir);
    }
    fs.watch(dir, function(eventType, filename) {
      if (!/___jb_(tmp|old)___$/.test(filename)) {
        console.log(filename.green);
        child.kill("SIGINT");
      }
    });
  }
}

dirs.forEach(recursive);

run();
