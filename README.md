# go-mongo-line


- Decription: use golang to implement golang line bot three function 
1. receive message (webhook) 
2. send message 
3. show user list

- test step: 
1. clone project and run "make local-debug"
2. To test receive message, add linebot "@134jxulb" and send message to this bot
3. To test send message, call api "send message" provided in Postmen. Line bot should send message to your account.
4. To get userList, call api "get all user" provided in Postmen. Service should return json with all user_id recorded in system.

- Postmen 
https://documenter.getpostman.com/view/1088678/2s935pr3wR 
