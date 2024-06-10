const { MongoClient } = require("mongodb");

let client;

const initializeMongoDB = async () => {
  const url =
    process.env.MONGODB_URL ||
    "mongodb://127.0.0.1:27017,127.0.0.1:27018,127.0.0.1:27019/?replicaSet=rs0";
  client = new MongoClient(url);
  console.log(`Connecting to MongoDB at ${url}`);
  await client.connect();
  console.log("Connected to MongoDB");
};

const getMongoClient = () => {
  return client;
};

module.exports = { initializeMongoDB, getMongoClient };
