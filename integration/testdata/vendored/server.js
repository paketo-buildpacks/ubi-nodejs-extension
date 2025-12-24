const http = require("http");
const port = process.env.PORT || 8080;
const bcrypt = require("bcrypt");

const requestHandler = (request, response) => {
  bcrypt.hash("Hello, World!", 10, function (err, hash) {
    if (err) {
      return console.error("Error hashing:", err);
    }
  });

  response.end("Hello, World!");
};

const server = http.createServer(requestHandler);

server.listen(port, (err) => {
  if (err) {
    return console.log("something bad happened", err);
  }

  console.log(`vendored server is listening on ${port}`);
});
