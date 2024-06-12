const WebSocket = require("ws");
const { getMongoClient } = require("./database");

const sockets = [];

const initializeWebsocketServer = (server) => {
  const websocketServer = new WebSocket.Server({ server });
  websocketServer.on("connection", onConnection);
  setInterval(broadcastPrices, 500);
};

const broadcastPrices = async () => {
  if (sockets.length === 0) return;
  const prices = await getLatestPrices();
  if (prices.length === 0) return;
  sockets.forEach((socket) => sendPrices(socket, prices));
};

const onConnection = async (ws) => {
  console.log("New websocket connection");
  sockets.push(ws);
  ws.on("close", () => onDisconnect(ws));
};

const getLatestPrices = async () => {
  const client = getMongoClient();
  const dbName = process.env.MONGODB_DB || "stockmarket";
  const collectionName = process.env.MONGODB_COLLECTION || "stocks";
  const result = await client
    .db(dbName)
    .collection(collectionName)
    .aggregate([
      {
        $sort: {
          _id: -1,
        },
      },
      {
        $group: {
          _id: "$company",
          latestEntry: { $first: "$$ROOT" },
        },
      },
    ])
    .toArray();
  const companies = result.map((entry) => entry.latestEntry);
  companies.sort((a, b) => a.company.localeCompare(b.company));
  return companies;
};

const sendPrices = async (ws, prices) => {
  ws.send(JSON.stringify({ type: "prices", prices }));
};

const onDisconnect = (ws) => {
  console.log("Websocket disconnected");
  sockets.splice(sockets.indexOf(ws), 1);
  console.log("Remaining sockets:", sockets.length);
};

module.exports = { initializeWebsocketServer };
