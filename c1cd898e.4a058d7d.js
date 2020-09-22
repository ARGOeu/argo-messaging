(window.webpackJsonp=window.webpackJsonp||[]).push([[16],{71:function(e,n,s){"use strict";s.r(n),s.d(n,"frontMatter",(function(){return c})),s.d(n,"metadata",(function(){return i})),s.d(n,"rightToc",(function(){return o})),s.d(n,"default",(function(){return b}));var t=s(2),r=s(6),a=(s(0),s(83)),c={id:"api_users",title:"User Management"},i={unversionedId:"api_users",id:"api_users",isDocsHomePage:!1,title:"User Management",description:"ARGO Messaging Service supports calls for creating and modifing users",source:"@site/docs/api_users.md",permalink:"/argo-messaging/docs/api_users",sidebar:"someSidebar",previous:{title:"Authentication",permalink:"/argo-messaging/docs/api_auth"},next:{title:"Projects",permalink:"/argo-messaging/docs/api_projects"}},o=[{value:"GET Manage Users - List all users",id:"get-manage-users---list-all-users",children:[{value:"Request",id:"request",children:[]},{value:"Paginated Request that returns all users in one page",id:"paginated-request-that-returns-all-users-in-one-page",children:[]},{value:"Example request",id:"example-request",children:[]},{value:"Responses",id:"responses",children:[]},{value:"Paginated Request that returns the 2 most recent users",id:"paginated-request-that-returns-the-2-most-recent-users",children:[]},{value:"Example request",id:"example-request-1",children:[]},{value:"Responses",id:"responses-1",children:[]},{value:"Paginated Request that returns the next 3 users",id:"paginated-request-that-returns-the-next-3-users",children:[]},{value:"Example request",id:"example-request-2",children:[]},{value:"Responses",id:"responses-2",children:[]},{value:"Paginated Request that returns all users that are members of a specific project",id:"paginated-request-that-returns-all-users-that-are-members-of-a-specific-project",children:[]},{value:"Example request",id:"example-request-3",children:[]},{value:"Responses",id:"responses-3",children:[]},{value:"Errors",id:"errors",children:[]}]},{value:"GET Manage Users - List a specific user",id:"get-manage-users---list-a-specific-user",children:[{value:"Request",id:"request-1",children:[]},{value:"Where",id:"where",children:[]},{value:"Example request",id:"example-request-4",children:[]},{value:"Responses",id:"responses-4",children:[]},{value:"Errors",id:"errors-1",children:[]}]},{value:"GET Manage Users - List a specific user by token",id:"get-manage-users---list-a-specific-user-by-token",children:[{value:"Request",id:"request-2",children:[]},{value:"Where",id:"where-1",children:[]},{value:"Example request",id:"example-request-5",children:[]},{value:"Responses",id:"responses-5",children:[]},{value:"Errors",id:"errors-2",children:[]}]},{value:"GET Manage Users - List a specific user by authentication key",id:"get-manage-users---list-a-specific-user-by-authentication-key",children:[{value:"Request",id:"request-3",children:[]},{value:"Example request",id:"example-request-6",children:[]},{value:"Responses",id:"responses-6",children:[]},{value:"Errors",id:"errors-3",children:[]}]},{value:"GET Manage Users - List a specific user by UUID",id:"get-manage-users---list-a-specific-user-by-uuid",children:[{value:"Request",id:"request-4",children:[]},{value:"Where",id:"where-2",children:[]},{value:"Example request",id:"example-request-7",children:[]},{value:"Responses",id:"responses-7",children:[]},{value:"Errors",id:"errors-4",children:[]}]},{value:"POST Manage Users - Create new user",id:"post-manage-users---create-new-user",children:[{value:"Request",id:"request-5",children:[]},{value:"Post body:",id:"post-body",children:[]},{value:"Where",id:"where-3",children:[]},{value:"Example request",id:"example-request-8",children:[]},{value:"Responses",id:"responses-8",children:[]},{value:"Errors",id:"errors-5",children:[]}]},{value:"PUT Manage Users - Update a user",id:"put-manage-users---update-a-user",children:[{value:"Request",id:"request-6",children:[]},{value:"Put body:",id:"put-body",children:[]},{value:"Where",id:"where-4",children:[]},{value:"Example request",id:"example-request-9",children:[]},{value:"Responses",id:"responses-9",children:[]},{value:"Errors",id:"errors-6",children:[]}]},{value:"POST Manage Users - Refresh token",id:"post-manage-users---refresh-token",children:[{value:"Request",id:"request-7",children:[]},{value:"Where",id:"where-5",children:[]},{value:"Example request",id:"example-request-10",children:[]},{value:"Responses",id:"responses-10",children:[]},{value:"Errors",id:"errors-7",children:[]}]},{value:"DELETE Manage Users - Delete User",id:"delete-manage-users---delete-user",children:[{value:"Request",id:"request-8",children:[]},{value:"Where",id:"where-6",children:[]},{value:"Example request",id:"example-request-11",children:[]},{value:"Responses",id:"responses-11",children:[]},{value:"Errors",id:"errors-8",children:[]}]}],l={rightToc:o};function b(e){var n=e.components,s=Object(r.a)(e,["components"]);return Object(a.b)("wrapper",Object(t.a)({},l,s,{components:n,mdxType:"MDXLayout"}),Object(a.b)("p",null,"ARGO Messaging Service supports calls for creating and modifing users"),Object(a.b)("h2",{id:"get-manage-users---list-all-users"},"[GET]"," Manage Users - List all users"),Object(a.b)("p",null,"This request lists all available users in the service using pagination"),Object(a.b)("p",null,"It is important to note that if there are no results to return the service will return the following:"),Object(a.b)("p",null,"Success Response\n",Object(a.b)("inlineCode",{parentName:"p"},"200 OK")),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{className:"language-json"}),'{\n "users": [],\n  "nextPageToken": "",\n  "totalSize": 0\n }\n')),Object(a.b)("p",null,"Also the default value for ",Object(a.b)("inlineCode",{parentName:"p"},"pageSize = 0")," and ",Object(a.b)("inlineCode",{parentName:"p"},'pageToken = "'),"."),Object(a.b)("p",null,Object(a.b)("inlineCode",{parentName:"p"},"Pagesize = 0")," returns all the results."),Object(a.b)("h3",{id:"request"},"Request"),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{className:"language-json"}),'GET "/v1/users"\n')),Object(a.b)("h3",{id:"paginated-request-that-returns-all-users-in-one-page"},"Paginated Request that returns all users in one page"),Object(a.b)("h3",{id:"example-request"},"Example request"),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{}),'curl -X GET -H "Content-Type: application/json"\n  "https://{URL}/v1/users?key=S3CR3T"\n')),Object(a.b)("h3",{id:"responses"},"Responses"),Object(a.b)("p",null,"If successful, the response contains a list of all available users in the service"),Object(a.b)("p",null,"Success Response\n",Object(a.b)("inlineCode",{parentName:"p"},"200 OK")),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{className:"language-json"}),'{\n "users": [\n    {\n       "uuid": "99bfd746-4ebe-11e8-9c2d-fa7ae01bbebw",\n       "projects": [\n          {\n             "project": "ARGO2",\n             "roles": [\n                "consumer",\n                "publisher"\n             ],\n             "topics": [],\n             "subscriptions": []\n          }\n       ],\n       "name": "Test",\n       "token": "S3CR3T",\n       "email": "Test@test.com",\n       "service_roles": [],\n       "created_on": "2009-11-10T23:00:00Z",\n       "modified_on": "2009-11-10T23:00:00Z"\n    },\n    {\n       "uuid": "99bfd746-4ebe-11e8-9c2d-fa7ae01bbebc",\n       "projects": [\n          {\n             "project": "ARGO",\n             "roles": [\n                "consumer",\n                "publisher"\n             ],\n             "topics": [\n                "topic1",\n                "topic2"\n             ],\n             "subscriptions": [\n                "sub1",\n                "sub2",\n                "sub3"\n             ]\n          }\n       ],\n       "name": "UserA",\n       "first_name": "FirstA",\n       "last_name": "LastA",\n       "organization": "OrgA",\n       "description": "DescA",\n       "token": "S3CR3T1",\n       "email": "foo-email",\n       "service_roles": [],\n       "created_on": "2009-11-10T23:00:00Z",\n       "modified_on": "2009-11-10T23:00:00Z"\n    },\n    {\n       "uuid": "94bfd746-4ebe-11e8-9c2d-fa7ae01bbebc",\n       "projects": [\n          {\n             "project": "ARGO",\n             "roles": [\n                "consumer",\n                "publisher"\n             ],\n             "topics": [\n                "topic1",\n                "topic2"\n             ],\n             "subscriptions": [\n                "sub1",\n                "sub3",\n                "sub4"\n             ]\n          }\n       ],\n       "name": "UserB",\n       "token": "S3CR3T2",\n       "email": "foo-email",\n       "service_roles": [],\n       "created_on": "2009-11-10T23:00:00Z",\n       "modified_on": "2009-11-10T23:00:00Z",\n       "created_by": "UserA"\n    },\n    {\n       "uuid": "99bfd746-4ebe-11e8-9c2d-fa7ae01bberr",\n       "projects": [\n          {\n             "project": "ARGO",\n             "roles": [\n                "publisher",\n                "consumer"\n             ],\n             "topics": [\n                "topic3"\n             ],\n             "subscriptions": [\n                "sub2"\n             ]\n          }\n       ],\n       "name": "UserX",\n       "token": "S3CR3T3",\n       "email": "foo-email",\n       "service_roles": [],\n       "created_on": "2009-11-10T23:00:00Z",\n       "modified_on": "2009-11-10T23:00:00Z",\n       "created_by": "UserA"\n    },\n    {\n       "uuid": "99bfd746-4ebe-11e8-9c2d-fa7ae01bbfrt",\n       "projects": [\n          {\n             "project": "ARGO",\n             "roles": [\n                "publisher",\n                "consumer"\n             ],\n             "topics": [\n                "topic2"\n             ],\n             "subscriptions": [\n                "sub3",\n                "sub4"\n             ]\n          }\n       ],\n       "name": "UserZ",\n       "token": "S3CR3T4",\n       "email": "foo-email",\n       "service_roles": [],\n       "created_on": "2009-11-10T23:00:00Z",\n       "modified_on": "2009-11-10T23:00:00Z",\n       "created_by": "UserA"\n    }\n ],\n "nextPageToken": "",\n "totalSize": 5\n}\n')),Object(a.b)("h3",{id:"paginated-request-that-returns-the-2-most-recent-users"},"Paginated Request that returns the 2 most recent users"),Object(a.b)("h3",{id:"example-request-1"},"Example request"),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{}),'curl -X GET -H "Content-Type: application/json"\n  "https://{URL}/v1/users?key=S3CR3T&pageSize=2"\n')),Object(a.b)("h3",{id:"responses-1"},"Responses"),Object(a.b)("p",null,"If successful, the response contains a list of the 2 most recently added users"),Object(a.b)("p",null,"Success Response\n",Object(a.b)("inlineCode",{parentName:"p"},"200 OK")),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{className:"language-json"}),'{\n "users": [\n    {\n       "uuid": "99bfd746-4ebe-11e8-9c2d-fa7ae01bbebw",\n       "projects": [\n          {\n             "project": "ARGO2",\n             "roles": [\n                "consumer",\n                "publisher"\n             ],\n             "topics": [],\n             "subscriptions": []\n          }\n       ],\n       "name": "Test",\n       "token": "S3CR3T",\n       "email": "Test@test.com",\n       "service_roles": [],\n       "created_on": "2009-11-10T23:00:00Z",\n       "modified_on": "2009-11-10T23:00:00Z"\n    },\n    {\n       "uuid": "99bfd746-4ebe-11e8-9c2d-fa7ae01bbebc",\n       "projects": [\n          {\n             "project": "ARGO",\n             "roles": [\n                "consumer",\n                "publisher"\n             ],\n             "topics": [\n                "topic1",\n                "topic2"\n             ],\n             "subscriptions": [\n                "sub1",\n                "sub2",\n                "sub3"\n             ]\n          }\n       ],\n       "name": "UserA",\n       "token": "S3CR3T1",\n       "email": "foo-email",\n       "service_roles": [],\n       "created_on": "2009-11-10T23:00:00Z",\n       "modified_on": "2009-11-10T23:00:00Z"\n    }\n ],\n "nextPageToken": "some_token2",\n "totalSize": 5\n}\n')),Object(a.b)("h3",{id:"paginated-request-that-returns-the-next-3-users"},"Paginated Request that returns the next 3 users"),Object(a.b)("h3",{id:"example-request-2"},"Example request"),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{}),'curl -X GET -H "Content-Type: application/json"\n  "https://{URL}/v1/users?key=S3CR3T&pageSize=3&pageToken=some_token2"\n')),Object(a.b)("h3",{id:"responses-2"},"Responses"),Object(a.b)("p",null,"If successful, the response contains a list of the next 3 users"),Object(a.b)("p",null,"Success Response\n",Object(a.b)("inlineCode",{parentName:"p"},"200 OK")),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{className:"language-json"}),'{\n "users": [\n    {\n       "uuid": "94bfd746-4ebe-11e8-9c2d-fa7ae01bbebc",\n       "projects": [\n          {\n             "project": "ARGO",\n             "roles": [\n                "consumer",\n                "publisher"\n             ],\n             "topics": [\n                "topic1",\n                "topic2"\n             ],\n             "subscriptions": [\n                "sub1",\n                "sub3",\n                "sub4"\n             ]\n          }\n       ],\n       "name": "UserB",\n       "token": "S3CR3T2",\n       "email": "foo-email",\n       "service_roles": [],\n       "created_on": "2009-11-10T23:00:00Z",\n       "modified_on": "2009-11-10T23:00:00Z",\n       "created_by": "UserA"\n    },\n    {\n       "uuid": "99bfd746-4ebe-11e8-9c2d-fa7ae01bberr",\n       "projects": [\n          {\n             "project": "ARGO",\n             "roles": [\n                "publisher",\n                "consumer"\n             ],\n             "topics": [\n                "topic3"\n             ],\n             "subscriptions": [\n                "sub2"\n             ]\n          }\n       ],\n       "name": "UserX",\n       "token": "S3CR3T3",\n       "email": "foo-email",\n       "service_roles": [],\n       "created_on": "2009-11-10T23:00:00Z",\n       "modified_on": "2009-11-10T23:00:00Z",\n       "created_by": "UserA"\n    },\n    {\n       "uuid": "99bfd746-4ebe-11e8-9c2d-fa7ae01bbfrt",\n       "projects": [\n          {\n             "project": "ARGO",\n             "roles": [\n                "publisher",\n                "consumer"\n             ],\n             "topics": [\n                "topic2"\n             ],\n             "subscriptions": [\n                "sub3",\n                "sub4"\n             ]\n          }\n       ],\n       "name": "UserZ",\n       "token": "S3CR3T4",\n       "email": "foo-email",\n       "service_roles": [],\n       "created_on": "2009-11-10T23:00:00Z",\n       "modified_on": "2009-11-10T23:00:00Z",\n       "created_by": "UserA"\n    }\n ],\n  "nextPageToken": "some_token3",\n  "totalSize": 5\n}\n')),Object(a.b)("h3",{id:"paginated-request-that-returns-all-users-that-are-members-of-a-specific-project"},"Paginated Request that returns all users that are members of a specific project"),Object(a.b)("h3",{id:"example-request-3"},"Example request"),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{}),'curl -X GET -H "Content-Type: application/json"\n  "https://{URL}/v1/users?key=S3CR3T&project=ARGO2"\n')),Object(a.b)("h3",{id:"responses-3"},"Responses"),Object(a.b)("p",null,"If successful, the response contains a list of all available users that are members in the project ARGO2"),Object(a.b)("p",null,"Success Response\n",Object(a.b)("inlineCode",{parentName:"p"},"200 OK")),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{className:"language-json"}),'{\n "users": [\n    {\n       "uuid": "99bfd746-4ebe-11e8-9c2d-fa7ae01bbebw",\n       "projects": [\n          {\n             "project": "ARGO2",\n             "roles": [\n                "consumer",\n                "publisher"\n             ],\n             "topics": [],\n             "subscriptions": []\n          }\n       ],\n       "name": "Test",\n       "token": "S3CR3T",\n       "email": "Test@test.com",\n       "service_roles": [],\n       "created_on": "2009-11-10T23:00:00Z",\n       "modified_on": "2009-11-10T23:00:00Z"\n    }\n ],\n "nextPageToken": "",\n "totalSize": 1\n}\n')),Object(a.b)("h3",{id:"errors"},"Errors"),Object(a.b)("p",null,"Please refer to section ",Object(a.b)("a",Object(t.a)({parentName:"p"},{href:"/argo-messaging/docs/api_errors"}),"Errors")," to see all possible Errors"),Object(a.b)("h2",{id:"get-manage-users---list-a-specific-user"},"[GET]"," Manage Users - List a specific user"),Object(a.b)("p",null,"This request lists information about a specific user in the service"),Object(a.b)("h3",{id:"request-1"},"Request"),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{}),'GET "/v1/users/{user_name}"\n')),Object(a.b)("h3",{id:"where"},"Where"),Object(a.b)("ul",null,Object(a.b)("li",{parentName:"ul"},"user_name: Name of the user")),Object(a.b)("h3",{id:"example-request-4"},"Example request"),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{}),'curl -X GET -H "Content-Type: application/json"\n  "https://{URL}/v1/users/UserA?key=S3CR3T"\n')),Object(a.b)("h3",{id:"responses-4"},"Responses"),Object(a.b)("p",null,"If successful, the response contains information about the specific user"),Object(a.b)("p",null,"Success Response\n",Object(a.b)("inlineCode",{parentName:"p"},"200 OK")),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{className:"language-json"}),'{\n   "uuid": "99bfd746-4rte-11e8-9c2d-fa7ae01bbebc",\n   "projects": [\n      {\n         "project": "ARGO",\n         "roles": [\n            "consumer",\n            "publisher"\n         ],\n         "topics": [\n            "topic1",\n            "topic2"\n         ],\n         "subscriptions": [\n            "sub1",\n            "sub2",\n            "sub3"\n         ]\n      }\n   ],\n   "name": "UserA",\n   "first_name": "FirstA",\n   "last_name": "LastA",\n   "organization": "OrgA",\n   "description": "DescA",\n   "token": "S3CR3T1",\n   "email": "foo-email",\n   "service_roles": [],\n   "created_on": "2009-11-10T23:00:00Z",\n   "modified_on": "2009-11-10T23:00:00Z"\n}\n')),Object(a.b)("h3",{id:"errors-1"},"Errors"),Object(a.b)("p",null,"Please refer to section ",Object(a.b)("a",Object(t.a)({parentName:"p"},{href:"/argo-messaging/docs/api_errors"}),"Errors")," to see all possible Errors"),Object(a.b)("h2",{id:"get-manage-users---list-a-specific-user-by-token"},"[GET]"," Manage Users - List a specific user by token"),Object(a.b)("p",null,"This request lists information about a specific user using user's token as input"),Object(a.b)("h3",{id:"request-2"},"Request"),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{}),'GET "/v1/users:byToken/{token}"\n')),Object(a.b)("h3",{id:"where-1"},"Where"),Object(a.b)("ul",null,Object(a.b)("li",{parentName:"ul"},"token: the token of the user")),Object(a.b)("h3",{id:"example-request-5"},"Example request"),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{}),'curl -X GET -H "Content-Type: application/json"\n  "https://{URL}/v1/users:byToken/S3CR3T1?key=S3CR3T"\n')),Object(a.b)("h3",{id:"responses-5"},"Responses"),Object(a.b)("p",null,"If successful, the response contains information about the specific user"),Object(a.b)("p",null,"Success Response\n",Object(a.b)("inlineCode",{parentName:"p"},"200 OK")),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{className:"language-json"}),'{\n   "uuid": "99bfd746-4ebe-11e8-9c2d-fa7ae01bbebc",\n   "projects": [\n      {\n         "project": "ARGO",\n         "roles": [\n            "consumer",\n            "publisher"\n         ],\n         "topics": [\n            "topic1",\n            "topic2"\n         ],\n         "subscriptions": [\n            "sub1",\n            "sub2",\n            "sub3"\n         ]\n      }\n   ],\n   "name": "UserA",\n   "first_name": "FirstA",\n   "last_name": "LastA",\n   "organization": "OrgA",\n   "description": "DescA",\n   "token": "S3CR3T1",\n   "email": "foo-email",\n   "service_roles": [],\n   "created_on": "2009-11-10T23:00:00Z",\n   "modified_on": "2009-11-10T23:00:00Z"\n}\n')),Object(a.b)("h3",{id:"errors-2"},"Errors"),Object(a.b)("p",null,"Please refer to section ",Object(a.b)("a",Object(t.a)({parentName:"p"},{href:"/argo-messaging/docs/api_errors"}),"Errors")," to see all possible Errors"),Object(a.b)("h2",{id:"get-manage-users---list-a-specific-user-by-authentication-key"},"[GET]"," Manage Users - List a specific user by authentication key"),Object(a.b)("p",null,"This request lists information about a specific user\nbased on the authentication key provided as a url parameter"),Object(a.b)("h3",{id:"request-3"},"Request"),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{}),'GET "/v1/users/profile"\n')),Object(a.b)("h3",{id:"example-request-6"},"Example request"),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{}),'curl -X GET -H "Content-Type: application/json"\n  "https://{URL}/v1/users/profile?key=S3CR3T1"\n')),Object(a.b)("h3",{id:"responses-6"},"Responses"),Object(a.b)("p",null,"If successful, the response contains information about the specific user"),Object(a.b)("p",null,"Success Response\n",Object(a.b)("inlineCode",{parentName:"p"},"200 OK")),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{className:"language-json"}),'{\n   "uuid": "99bfd746-4ebe-11e8-9c2d-fa7ae01bbebc",\n   "projects": [\n      {\n         "project": "ARGO",\n         "roles": [\n            "consumer",\n            "publisher"\n         ],\n         "topics": [\n            "topic1",\n            "topic2"\n         ],\n         "subscriptions": [\n            "sub1",\n            "sub2",\n            "sub3"\n         ]\n      }\n   ],\n   "name": "UserA",\n   "first_name": "FirstA",\n   "last_name": "LastA",\n   "organization": "OrgA",\n   "description": "DescA",\n   "token": "S3CR3T1",\n   "email": "foo-email",\n   "service_roles": [],\n   "created_on": "2009-11-10T23:00:00Z",\n   "modified_on": "2009-11-10T23:00:00Z"\n}\n')),Object(a.b)("h3",{id:"errors-3"},"Errors"),Object(a.b)("p",null,"Please refer to section ",Object(a.b)("a",Object(t.a)({parentName:"p"},{href:"/argo-messaging/docs/api_errors"}),"Errors")," to see all possible Errors"),Object(a.b)("h2",{id:"get-manage-users---list-a-specific-user-by-uuid"},"[GET]"," Manage Users - List a specific user by UUID"),Object(a.b)("p",null,"This request lists information about a specific user using user's UUID as input"),Object(a.b)("h3",{id:"request-4"},"Request"),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{}),'GET "/v1/users:byUUID/{uuid}"\n')),Object(a.b)("h3",{id:"where-2"},"Where"),Object(a.b)("ul",null,Object(a.b)("li",{parentName:"ul"},"uuid: the uuid of the user")),Object(a.b)("h3",{id:"example-request-7"},"Example request"),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{}),'curl -X GET -H "Content-Type: application/json"\n  "https://{URL}/v1/users:byUUID/99bfd746-4ebe-11e8-9c2d-fa7ae01bbebc?key=S3CR3T"\n')),Object(a.b)("h3",{id:"responses-7"},"Responses"),Object(a.b)("p",null,"If successful, the response contains information about the specific user"),Object(a.b)("p",null,"Success Response\n",Object(a.b)("inlineCode",{parentName:"p"},"200 OK")),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{className:"language-json"}),'{\n   "uuid": "99bfd746-4ebe-11e8-9c2d-fa7ae01bbebc",\n   "projects": [\n      {\n         "project": "ARGO",\n         "roles": [\n            "consumer",\n            "publisher"\n         ],\n         "topics": [\n            "topic1",\n            "topic2"\n         ],\n         "subscriptions": [\n            "sub1",\n            "sub2",\n            "sub3"\n         ]\n      }\n   ],\n   "name": "UserA",\n   "first_name": "FirstA",\n   "last_name": "LastA",\n   "organization": "OrgA",\n   "description": "DescA",\n   "token": "S3CR3T1",\n   "email": "foo-email",\n   "service_roles": [],\n   "created_on": "2009-11-10T23:00:00Z",\n   "modified_on": "2009-11-10T23:00:00Z"\n}\n')),Object(a.b)("h3",{id:"errors-4"},"Errors"),Object(a.b)("p",null,"Please refer to section ",Object(a.b)("a",Object(t.a)({parentName:"p"},{href:"/argo-messaging/docs/api_errors"}),"Errors")," to see all possible Errors"),Object(a.b)("h2",{id:"post-manage-users---create-new-user"},"[POST]"," Manage Users - Create new user"),Object(a.b)("p",null,"This request creates a new user in a project"),Object(a.b)("h3",{id:"request-5"},"Request"),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{className:"language-json"}),'POST "/v1/users/{user_name}"\n')),Object(a.b)("h3",{id:"post-body"},"Post body:"),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{className:"language-json"}),'{\n "projects": [\n    {\n       "project": "ARGO",\n       "roles": [\n          "project_admin"\n       ]\n    }\n ],\n "email": "foo-email",\n "first_name": "fname-1",\n "last_name": "lname-1",\n "organization": "org-1",\n "description": "desc-1",\n "service_roles":[]\n}\n')),Object(a.b)("h3",{id:"where-3"},"Where"),Object(a.b)("ul",null,Object(a.b)("li",{parentName:"ul"},"user_name: Name of the user"),Object(a.b)("li",{parentName:"ul"},"projects: A list of Projects & associated roles that the user has on those projects"),Object(a.b)("li",{parentName:"ul"},"email: User's email"),Object(a.b)("li",{parentName:"ul"},"service_roles: A list of service-wide roles. An example of service-wide role is ",Object(a.b)("inlineCode",{parentName:"li"},"service_admin")," which can manage projects or other users")),Object(a.b)("h5",{id:"available-roles"},"Available Roles"),Object(a.b)("p",null,"ARGO Messaging Service has the following predefined project roles:"),Object(a.b)("table",null,Object(a.b)("thead",{parentName:"table"},Object(a.b)("tr",{parentName:"thead"},Object(a.b)("th",Object(t.a)({parentName:"tr"},{align:null}),"Role"),Object(a.b)("th",Object(t.a)({parentName:"tr"},{align:null}),"Description"))),Object(a.b)("tbody",{parentName:"table"},Object(a.b)("tr",{parentName:"tbody"},Object(a.b)("td",Object(t.a)({parentName:"tr"},{align:null}),"project_admin"),Object(a.b)("td",Object(t.a)({parentName:"tr"},{align:null}),"Users that have the ",Object(a.b)("inlineCode",{parentName:"td"},"project_admin")," have, by default, all capabilities in their project. They can also manage resources such as topics and subscriptions (CRUD) and also manage ACLs (users) on those resources as well")),Object(a.b)("tr",{parentName:"tbody"},Object(a.b)("td",Object(t.a)({parentName:"tr"},{align:null}),"consumer"),Object(a.b)("td",Object(t.a)({parentName:"tr"},{align:null}),"Users that have the ",Object(a.b)("inlineCode",{parentName:"td"},"consumer")," role are only able to pull messages from subscriptions that are authorized to use (based on ACLs)")),Object(a.b)("tr",{parentName:"tbody"},Object(a.b)("td",Object(t.a)({parentName:"tr"},{align:null}),"publisher"),Object(a.b)("td",Object(t.a)({parentName:"tr"},{align:null}),"Users that have the ",Object(a.b)("inlineCode",{parentName:"td"},"publisher")," role are only able to publish messages on topics that are authorized to use (based on ACLs)")))),Object(a.b)("p",null,"and the following service-wide role:"),Object(a.b)("table",null,Object(a.b)("thead",{parentName:"table"},Object(a.b)("tr",{parentName:"thead"},Object(a.b)("th",Object(t.a)({parentName:"tr"},{align:null}),"Role"),Object(a.b)("th",Object(t.a)({parentName:"tr"},{align:null}),"Description"))),Object(a.b)("tbody",{parentName:"table"},Object(a.b)("tr",{parentName:"tbody"},Object(a.b)("td",Object(t.a)({parentName:"tr"},{align:null}),"service_admin"),Object(a.b)("td",Object(t.a)({parentName:"tr"},{align:null}),"Users with ",Object(a.b)("inlineCode",{parentName:"td"},"service_admin")," role operate service wide. They are able to create, modify and delete projects. Also they are able to create, modify and delete users and assign them to projects.")))),Object(a.b)("h3",{id:"example-request-8"},"Example request"),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{}),'json\ncurl -X POST -H "Content-Type: application/json"\n -d POSTDATA "https://{URL}/v1/projects/ARGO/users/USERNEW?key=S3CR3T"\n')),Object(a.b)("h3",{id:"responses-8"},"Responses"),Object(a.b)("p",null,"If successful, the response contains the newly created user"),Object(a.b)("p",null,"Success Response\n",Object(a.b)("inlineCode",{parentName:"p"},"200 OK")),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{className:"language-json"}),'{\n "uuid": "99bfd746-4ebe-11e8-9c2a-fa7ae01bbebc",\n "projects": [\n    {\n       "project": "ARGO",\n       "roles": [\n          "project_admin"\n       ],\n       "topics":[],\n       "subscriptions":[]\n    }\n ],\n "name": "USERNEW",\n "token": "R4ND0MT0K3N",\n "email": "foo-email",\n "first_name": "fname-1",\n "last_name": "lname-1",\n "organization": "org-1",\n "description": "desc-1",\n "service_roles":[],\n "created_on": "2009-11-10T23:00:00Z",\n "modified_on": "2009-11-10T23:00:00Z",\n "created_by": "UserA"\n}\n')),Object(a.b)("h3",{id:"errors-5"},"Errors"),Object(a.b)("p",null,"Please refer to section ",Object(a.b)("a",Object(t.a)({parentName:"p"},{href:"/argo-messaging/docs/api_errors"}),"Errors")," to see all possible Errors"),Object(a.b)("h2",{id:"put-manage-users---update-a-user"},"[PUT]"," Manage Users - Update a user"),Object(a.b)("p",null,"This request updates an existing user's information"),Object(a.b)("h3",{id:"request-6"},"Request"),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{className:"language-json"}),'PUT "/v1/users/{user_name}"\n')),Object(a.b)("h3",{id:"put-body"},"Put body:"),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{className:"language-json"}),'{\n"uuid": "99bfd746-4ebe-11e8-9c2d-fa7ae01bbebz",\n "projects": [\n    {\n       "project": "ARGO2",\n       "roles": [\n          "project_admin"\n       ]\n    }\n ],\n "name": "CHANGED_NAME",\n "first_name": "fname-1",\n "last_name": "lname-1",\n "organization": "org-1",\n "description": "desc-1",\n "email": "foo-email",\n "service_roles":[]\n}\n')),Object(a.b)("h3",{id:"where-4"},"Where"),Object(a.b)("ul",null,Object(a.b)("li",{parentName:"ul"},"user_name: Name of the user"),Object(a.b)("li",{parentName:"ul"},"projects: A list of Projects & associated roles that the user has on those projects"),Object(a.b)("li",{parentName:"ul"},"email: User's email"),Object(a.b)("li",{parentName:"ul"},"service_roles: A list of service-wide roles. An example of service-wide role is ",Object(a.b)("inlineCode",{parentName:"li"},"service_admin")," which can manage projects or other users")),Object(a.b)("h3",{id:"example-request-9"},"Example request"),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{}),'json\ncurl -X POST -H "Content-Type: application/json"\n -d PUTDATA "https://{URL}/v1/projects/ARGO/users/USERNEW?key=S3CR3T"\n')),Object(a.b)("h3",{id:"responses-9"},"Responses"),Object(a.b)("p",null,"If successful, the response contains the newly created project"),Object(a.b)("p",null,"Success Response\n",Object(a.b)("inlineCode",{parentName:"p"},"200 OK")),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{className:"language-json"}),'{\n"uuid": "99bfd740-4ebe-11e8-9c2d-fa7ae01bbebc",\n "projects": [\n    {\n       "project": "ARGO2",\n       "roles": [\n          "project_admin"\n       ],\n       "topics":[],\n       "subscriptions":[]\n    }\n ],\n "name": "CHANGED_NAME",\n "token": "R4ND0MT0K3N",\n "email": "foo-email",\n "first_name": "fname-1",\n "last_name": "lname-1",\n "organization": "org-1",\n "description": "desc-1",\n "service_roles":[],\n "created_on": "2009-11-10T23:00:00Z",\n "modified_on": "2009-11-11T10:00:00Z",\n "created_by": "UserA"\n}\n')),Object(a.b)("h3",{id:"errors-6"},"Errors"),Object(a.b)("p",null,"Please refer to section ",Object(a.b)("a",Object(t.a)({parentName:"p"},{href:"/argo-messaging/docs/api_errors"}),"Errors")," to see all possible Errors"),Object(a.b)("h2",{id:"post-manage-users---refresh-token"},"[POST]"," Manage Users - Refresh token"),Object(a.b)("p",null,"This request refreshes an existing user's token"),Object(a.b)("h3",{id:"request-7"},"Request"),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{className:"language-json"}),'POST "/v1/users/{user_name}:refreshToken"\n')),Object(a.b)("h3",{id:"where-5"},"Where"),Object(a.b)("ul",null,Object(a.b)("li",{parentName:"ul"},"user_name: Name of the user")),Object(a.b)("h3",{id:"example-request-10"},"Example request"),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{}),'json\ncurl -X POST -H "Content-Type: application/json"\n "https://{URL}/v1/projects/ARGO/users/USER2:refreshToken?key=S3CR3T"\n')),Object(a.b)("h3",{id:"responses-10"},"Responses"),Object(a.b)("p",null,"If successful, the response contains the newly created project"),Object(a.b)("p",null,"Success Response\n",Object(a.b)("inlineCode",{parentName:"p"},"200 OK")),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{className:"language-json"}),'{\n"uuid": "99bfd746-4ebe-11p0-9c2d-fa7ae01bbebc",\n "projects": [\n    {\n       "project": "ARGO",\n       "roles": [\n          "project_admin"\n       ],\n       "topics":[],\n       "subscriptions":[]\n    }\n ],\n "name": "USER2",\n "token": "NEWRANDOMTOKEN",\n "email": "foo-email",\n "service_roles":[],\n "created_on": "2009-11-10T23:00:00Z",\n "modified_on": "2009-11-11T12:00:00Z",\n "created_by": "UserA"\n}\n')),Object(a.b)("h3",{id:"errors-7"},"Errors"),Object(a.b)("p",null,"Please refer to section ",Object(a.b)("a",Object(t.a)({parentName:"p"},{href:"/argo-messaging/docs/api_errors"}),"Errors")," to see all possible Errors"),Object(a.b)("h2",{id:"delete-manage-users---delete-user"},"[DELETE]"," Manage Users - Delete User"),Object(a.b)("p",null,"This request deletes an existing user"),Object(a.b)("h3",{id:"request-8"},"Request"),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{className:"language-json"}),'DELETE "/v1/users/{user_name}"\n')),Object(a.b)("h3",{id:"where-6"},"Where"),Object(a.b)("ul",null,Object(a.b)("li",{parentName:"ul"},"user_name: Name of the user")),Object(a.b)("h3",{id:"example-request-11"},"Example request"),Object(a.b)("pre",null,Object(a.b)("code",Object(t.a)({parentName:"pre"},{className:"language-json"}),'curl -X DELETE -H "Content-Type: application/json"\n "https://{URL}/v1/projects/ARGO/users/USER2?key=S3CR3T"\n')),Object(a.b)("h3",{id:"responses-11"},"Responses"),Object(a.b)("p",null,"If successful, the response returns empty"),Object(a.b)("p",null,"Success Response\n",Object(a.b)("inlineCode",{parentName:"p"},"200 OK")),Object(a.b)("h3",{id:"errors-8"},"Errors"),Object(a.b)("p",null,"Please refer to section ",Object(a.b)("a",Object(t.a)({parentName:"p"},{href:"/argo-messaging/docs/api_errors"}),"Errors")," to see all possible Errors"))}b.isMDXComponent=!0},83:function(e,n,s){"use strict";s.d(n,"a",(function(){return u})),s.d(n,"b",(function(){return j}));var t=s(0),r=s.n(t);function a(e,n,s){return n in e?Object.defineProperty(e,n,{value:s,enumerable:!0,configurable:!0,writable:!0}):e[n]=s,e}function c(e,n){var s=Object.keys(e);if(Object.getOwnPropertySymbols){var t=Object.getOwnPropertySymbols(e);n&&(t=t.filter((function(n){return Object.getOwnPropertyDescriptor(e,n).enumerable}))),s.push.apply(s,t)}return s}function i(e){for(var n=1;n<arguments.length;n++){var s=null!=arguments[n]?arguments[n]:{};n%2?c(Object(s),!0).forEach((function(n){a(e,n,s[n])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(s)):c(Object(s)).forEach((function(n){Object.defineProperty(e,n,Object.getOwnPropertyDescriptor(s,n))}))}return e}function o(e,n){if(null==e)return{};var s,t,r=function(e,n){if(null==e)return{};var s,t,r={},a=Object.keys(e);for(t=0;t<a.length;t++)s=a[t],n.indexOf(s)>=0||(r[s]=e[s]);return r}(e,n);if(Object.getOwnPropertySymbols){var a=Object.getOwnPropertySymbols(e);for(t=0;t<a.length;t++)s=a[t],n.indexOf(s)>=0||Object.prototype.propertyIsEnumerable.call(e,s)&&(r[s]=e[s])}return r}var l=r.a.createContext({}),b=function(e){var n=r.a.useContext(l),s=n;return e&&(s="function"==typeof e?e(n):i(i({},n),e)),s},u=function(e){var n=b(e.components);return r.a.createElement(l.Provider,{value:n},e.children)},p={inlineCode:"code",wrapper:function(e){var n=e.children;return r.a.createElement(r.a.Fragment,{},n)}},d=r.a.forwardRef((function(e,n){var s=e.components,t=e.mdxType,a=e.originalType,c=e.parentName,l=o(e,["components","mdxType","originalType","parentName"]),u=b(s),d=t,j=u["".concat(c,".").concat(d)]||u[d]||p[d]||a;return s?r.a.createElement(j,i(i({ref:n},l),{},{components:s})):r.a.createElement(j,i({ref:n},l))}));function j(e,n){var s=arguments,t=n&&n.mdxType;if("string"==typeof e||t){var a=s.length,c=new Array(a);c[0]=d;var i={};for(var o in n)hasOwnProperty.call(n,o)&&(i[o]=n[o]);i.originalType=e,i.mdxType="string"==typeof e?e:t,c[1]=i;for(var l=2;l<a;l++)c[l]=s[l];return r.a.createElement.apply(null,c)}return r.a.createElement.apply(null,s)}d.displayName="MDXCreateElement"}}]);