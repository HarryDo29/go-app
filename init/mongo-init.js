db = db.getSiblingDB("go-app");

// tạo collection
db.createCollection("messages");

// tạo user riêng cho app
db.createUser({
  user: "user",
  pwd: "123456",
  roles: [
    {
      role: "readWrite",
      db: "go-app",
    },
  ],
});