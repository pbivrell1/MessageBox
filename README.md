# MessageBox
## overview
This repo contains everything needed to deploy the small message box api described in the QM challenge.\
The docker-compose file will deploy a redis database and go http server with the required API implementation. The redis deployment can be configured by editing the  /redis/conf/redis.conf file located in the repository before deploying with docker-compose

The server scaffolding and middleware for this project was generated using [oapi-codegen](https://github.com/deepmap/oapi-codegen). To obtain the swagger yaml necessary to generate this code, I used the base64 encoded yaml file found in the http request stream when loading the provided swagger page.

The provided test collection has been slightly modified to include 201 (rather than 202) status codes per the swagger spec. However, there are other problems with the test collection that are described in detail toward the bottom of this document.
## how to run
Pre-requisites: docker and docker-compose must be installed.\
To run this package once the pre-requisites are met:

* from the base repo directory run `docker-compose up --build`. This will deploy two containers: one for the redis database and another for the http server
* get the IP address of your docker container by running `docker ps` then identifying the `CONTAINER ID` of the `Messagebox` container. Then run `docker inspect [CONTAINER ID]` and identify the IP address of the container.

You should now be able to make requests to the API on `[CONTAINER IP ADDRESS]:3001`

## test collection problems and notes
A collection of postman tests can be found in the `/pkg/test` directory. These tests have been slightly modified from their original version. The only modifications were to the HTTP status codes tested against during successful message and reply POST requests. The swagger specification calls for 201 Status Created, but the tests have mis-matched wording referring to both 202 and 201 return codes. I opted to follow the swagger and make these 201s. 

After some detailed investigation and tracking the state through the provided test collection, I believe there are 7 tests that fail which violate the requirements specified in the challenge. The failing tests are the GET users/{username}/mailbox with replies tests. 

Here's a breakdown of the states reached after each post message/reply in the collection versus the failing tests' expectations at the end:

( each message is pushed to the front of the recipient mailbox )
* **post1**: user1 sends message(1) to a group that contains only user1
* **state1**: 
  * user1 mailbox: [0]message from a group


* **post2**: user1 sends message(2) to user2
* **state2**: 
  * user1 mailbox:  [0]message from a group 
  * user2 mailbox:  [0]message from user1


* **post3**: user1 sends reply to group message(1) - reply goes to the group (just user1) and user1 receives 1 copy of the message ( per requirements )
* **state3**: user1 mailbox:  [0]reply to group, [1]message from a group
user2 mailbox:  [0]message from user1


* **post4**: user1 replies to a message sent from user1 to user2 (requirements doc states we don't need to worry about this condition). Per the requirements, we send this message to user 1 since they were the original sender of the message being replied to.
* **state4**: 
  * user1 mailbox: [0]reply to message to user2, [1]reply to group , [2]message from a group
  * user2 mailbox: [0]message from user1

The failing tests expect the following:

user1 to have 2 messages - the message at [1] to be the original group message sent by user1 to its own group - seemingly user1 should have not received one of either the group reply or the reply sent to the message originally sent from user1 to user2 but this conflicts with the requirements.

user2 to have 2 messages - the message at [1] should be the message(2) sent from user 1, the message at [0] should be the reply message(4) sent by user1 to the message sent from user 1 to user 2. However, user 2 does not have this message since it was sent as a reply to a message originally sent by user1.

I've thought out a few ways to manipulate both the code and test logic to get some of these to pass. For instance, if I introduce logic to send a user to user reply to both the original sender and original recipient, user2's mailbox tests will pass since they receive the reply to message(2) (the message from user1 to user2).
