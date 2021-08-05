(window.webpackJsonp=window.webpackJsonp||[]).push([[5],{60:function(e,t,a){"use strict";a.r(t),a.d(t,"frontMatter",(function(){return i})),a.d(t,"metadata",(function(){return c})),a.d(t,"rightToc",(function(){return b})),a.d(t,"default",(function(){return o}));var n=a(2),s=a(6),r=(a(0),a(85)),i={id:"overview",title:"Overview & Introduction"},c={unversionedId:"overview",id:"overview",isDocsHomePage:!1,title:"Overview & Introduction",description:"The Messaging Services is implemented as a Publish/Subscribe Service. Instead of focusing on a single Messaging API specification for handling the logic of publishing/subscribing to the broker network the API focuses on creating nodes of Publishers and Subscribers as a Service.",source:"@site/docs/overview.md",permalink:"/argo-messaging/docs/overview",sidebar:"someSidebar",previous:{title:"AMS - The Service",permalink:"/argo-messaging/docs/"},next:{title:"Using Apache Kafka as a Backend Message system",permalink:"/argo-messaging/docs/msg_backend"}},b=[{value:"Terminology",id:"terminology",children:[]},{value:"The ARGO Messaging Service",id:"the-argo-messaging-service",children:[]},{value:"Topics",id:"topics",children:[]},{value:"Subscriptions",id:"subscriptions",children:[]},{value:"Pull vs Push Subscriptions",id:"pull-vs-push-subscriptions",children:[{value:"Pull subscriptions",id:"pull-subscriptions",children:[]},{value:"Push subscriptions",id:"push-subscriptions",children:[]}]},{value:"Messages",id:"messages",children:[]},{value:"Message acknowledgement deadline",id:"message-acknowledgement-deadline",children:[]}],l={rightToc:b};function o(e){var t=e.components,a=Object(s.a)(e,["components"]);return Object(r.b)("wrapper",Object(n.a)({},l,a,{components:t,mdxType:"MDXLayout"}),Object(r.b)("p",null,"The Messaging Services is implemented as a Publish/Subscribe Service. Instead of focusing on a single Messaging API specification for handling the logic of publishing/subscribing to the broker network the API focuses on creating nodes of Publishers and Subscribers as a Service."),Object(r.b)("h2",{id:"terminology"},"Terminology"),Object(r.b)("table",null,Object(r.b)("thead",{parentName:"table"},Object(r.b)("tr",{parentName:"thead"},Object(r.b)("th",Object(n.a)({parentName:"tr"},{align:null}),"Term"),Object(r.b)("th",Object(n.a)({parentName:"tr"},{align:null}),"Description"))),Object(r.b)("tbody",{parentName:"table"},Object(r.b)("tr",{parentName:"tbody"},Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"Project"),Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"ARGO Messaging Service uses the notion of \u2018projects\u2019 for each tenant and can support multiple projects each one containing multiple topics/subscriptions and users on the same Kafka backend.")),Object(r.b)("tr",{parentName:"tbody"},Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"topic"),Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"A named resource to which messages are sent by publishers. A topic name must be scoped to a project.")),Object(r.b)("tr",{parentName:"tbody"},Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"subscription"),Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"A named resource representing the stream of messages from a single, specific topic, to be delivered to the subscribing application. A subscription name  must be scoped to a project.")),Object(r.b)("tr",{parentName:"tbody"},Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"ackDeadlineSeconds"),Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"Time window in seconds during which client can send an acknowledgement to notify the Service that a message has been successfully received")),Object(r.b)("tr",{parentName:"tbody"},Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"ack"),Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"Acknowledgement issued by the client that the message has been received")),Object(r.b)("tr",{parentName:"tbody"},Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"pushConfig"),Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"Contains information about the push endpoint")),Object(r.b)("tr",{parentName:"tbody"},Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"pushEndpoint"),Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"Webhook URL which will receive the messages")),Object(r.b)("tr",{parentName:"tbody"},Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"Message"),Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"The combination of data and (optional) attributes that a publisher sends to a topic and is eventually delivered to subscribers.")),Object(r.b)("tr",{parentName:"tbody"},Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"Messages - messageId"),Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"Id of the message - GENERATED by the api")),Object(r.b)("tr",{parentName:"tbody"},Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"Messages - data"),Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"Data payload ALWAYS encoded in Base64")),Object(r.b)("tr",{parentName:"tbody"},Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"Messages - attributes"),Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"Dictionary with key/value metadata - OPTIONAL")),Object(r.b)("tr",{parentName:"tbody"},Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"Messages - publishTime"),Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"Timestamp when the message was published, in UTC Zulu time format - detailed to nanoseconds. - GENERATED BY THE API (UTC+2 at devel infrastructure)")),Object(r.b)("tr",{parentName:"tbody"},Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"AMS"),Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"ARGO Messaging Service")),Object(r.b)("tr",{parentName:"tbody"},Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"maxMessages"),Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"the max number of messages returned by one call by setting maxMessages field (when a client pull messages from a subscription).")),Object(r.b)("tr",{parentName:"tbody"},Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"returnImmediately"),Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"By default, the server will keep the connection open until at least one message is received; you can optionally set the returnImmediately field to true to prevent the subscriber from waiting if the queue is currently empty. (when a client pull messages from a subscription).")))),Object(r.b)("h2",{id:"the-argo-messaging-service"},"The ARGO Messaging Service"),Object(r.b)("p",null,"The ARGO Messaging Service is a Publish/Subscribe Service, which implements the Google PubSub protocol. It provides an HTTP API that enables Users/Systems to implement message oriented service using the Publish/Subscribe Model over plain HTTP."),Object(r.b)("p",null,"In the Publish/Subscribe paradigm, Publishers are users/systems that can send messages to named-channels called Topics. Subscribers are users/systems that create Subscriptions to specific topics and receive messages."),Object(r.b)("h2",{id:"topics"},"Topics"),Object(r.b)("p",null,"Topics are resources that can hold messages. Publishers (users/systems) can create topics on demand and name them (Usually with names that make sense and express the class of messages delivered in the topic)"),Object(r.b)("h2",{id:"subscriptions"},"Subscriptions"),Object(r.b)("p",null,"In order for a user to be able to consume messages, he must first create a subscription. Subscriptions are resources that can be created by users   on demand and are attached to specific topics. Each topic can have multiple subscriptions but each subscription can be attached to just one topic. Subscriptions allows Subscribers to incrementally consume messages, at their own pace, while the progress is automatically tracked for each subscription."),Object(r.b)("h2",{id:"pull-vs-push-subscriptions"},"Pull vs Push Subscriptions"),Object(r.b)("p",null,"Pub/Sub supports both push and pull message delivery. In push delivery, the Pub/Sub initiates requests to your subscriber application to deliver messages. In pull delivery, your subscription application initiates requests to the Pub/Sub server to retrieve messages."),Object(r.b)("h3",{id:"pull-subscriptions"},"Pull subscriptions"),Object(r.b)("p",null,"Pull subscriptions can be configured to require that message deliveries are acknowledged by the Subscribers. If an acknowledgement is made, subscription can resume progressing and send the next available messages. If no acknowledgement is made subscription pauses progressing and re-sends the same messages."),Object(r.b)("p",null,"In a pull subscription, the subscribing application explicitly calls the API pull method, which requests delivery of a message in the subscription queue. The Pub/Sub server responds with the message (or an error if the queue is empty), and an ack ID. The subscriber then explicitly calls the acknowledge method, using the returned ack ID, to acknowledge receipt."),Object(r.b)("h3",{id:"push-subscriptions"},"Push subscriptions"),Object(r.b)("p",null,'In a push subscription, the Pub/Sub server sends a request to the subscriber application, at a preconfigured endpoint. The subscriber\'s HTTP response serves as an implicit acknowledgement: a success response indicates that the message has been successfully processed and the Pub/Sub system can delete it from the subscription; a non-success response indicates that the Pub/Sub server should resend it (implicit "nack"). To ensure that subscribers can handle the message flow, the Pub/Sub dynamically adjusts the flow of requests and uses an algorithm to rate-limit retries.'),Object(r.b)("blockquote",null,Object(r.b)("p",{parentName:"blockquote"},"In the current implementation of the AMS there is support only for pull subscriptions.\nSupport for push subscriptions will be available in a later version.")),Object(r.b)("h2",{id:"messages"},"Messages"),Object(r.b)("p",null,"In the ARGO Messaging Service each message has an identifier, data (payload) and metadata (optional). The metadata are stored in a attribute dictionary as key/value pairs. The message is represented in json format as follows:"),Object(r.b)("pre",null,Object(r.b)("code",Object(n.a)({parentName:"pre"},{className:"language-json"}),' {\n   "messageId": "12",\n   "data": "YmFzZTY0",\n   "attributes": [\n     {\n       "key": "attribute1",\n       "value": "value1"\n     },\n     {\n       "key": "attribute2",\n       "value": "value2"\n     }\n   ],\n   "publishTime":"2016-03-15T17:11:34.035345612Z"  \n }\n')),Object(r.b)("table",null,Object(r.b)("thead",{parentName:"table"},Object(r.b)("tr",{parentName:"thead"},Object(r.b)("th",Object(n.a)({parentName:"tr"},{align:null}),"Field"),Object(r.b)("th",Object(n.a)({parentName:"tr"},{align:null}),"Description"))),Object(r.b)("tbody",{parentName:"table"},Object(r.b)("tr",{parentName:"tbody"},Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"messageId"),Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"Id of the message - GENERATED by the ARGO Messaging Service. Judging from interaction with the service emulator locally and with the service itself online, yes the messages were identified with sequential numbers.")),Object(r.b)("tr",{parentName:"tbody"},Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"data"),Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"Data payload ALWAYS encoded in Base64")),Object(r.b)("tr",{parentName:"tbody"},Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"attributes"),Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"Dictionary with key/value metadata - OPTIONAL")),Object(r.b)("tr",{parentName:"tbody"},Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"publishTime"),Object(r.b)("td",Object(n.a)({parentName:"tr"},{align:null}),"Timestamp when the message was published, in UTC Zulu time format - detailed to nanoseconds. - GENERATED BY THE API")))),Object(r.b)("h2",{id:"message-acknowledgement-deadline"},"Message acknowledgement deadline"),Object(r.b)("p",null,"The ack deadline is the number of seconds after delivery, during which the subscriber must acknowledge the receipt of a pull or push message. If a subscriber does not respond with an explicit acknowledge (for a pull subscriber) or with a success response code (for a push subscriber) by this deadline, the server will attempt to resend the message. By default this deadline is 10 seconds."),Object(r.b)("p",null,"If a client tries to acknowledge a message while the Ack period has passed it will receive a 408 ERROR in the following format:"),Object(r.b)("pre",null,Object(r.b)("code",Object(n.a)({parentName:"pre"},{className:"language-json"}),'{\n  "error": {\n    "code": 408,\n    "message": "ack timeout",\n    "status": "TIMEOUT"\n  }\n}\n')),Object(r.b)("p",null,"The Ack deadline can be set-up to a higher number during subscription creation by assigning a value to ",Object(r.b)("inlineCode",{parentName:"p"},"ackDeadlineSeconds")," json field. More on subscription creation ",Object(r.b)("a",Object(n.a)({parentName:"p"},{href:"/argo-messaging/docs/api_subscriptions"}),"here")))}o.isMDXComponent=!0},85:function(e,t,a){"use strict";a.d(t,"a",(function(){return u})),a.d(t,"b",(function(){return m}));var n=a(0),s=a.n(n);function r(e,t,a){return t in e?Object.defineProperty(e,t,{value:a,enumerable:!0,configurable:!0,writable:!0}):e[t]=a,e}function i(e,t){var a=Object.keys(e);if(Object.getOwnPropertySymbols){var n=Object.getOwnPropertySymbols(e);t&&(n=n.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),a.push.apply(a,n)}return a}function c(e){for(var t=1;t<arguments.length;t++){var a=null!=arguments[t]?arguments[t]:{};t%2?i(Object(a),!0).forEach((function(t){r(e,t,a[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(a)):i(Object(a)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(a,t))}))}return e}function b(e,t){if(null==e)return{};var a,n,s=function(e,t){if(null==e)return{};var a,n,s={},r=Object.keys(e);for(n=0;n<r.length;n++)a=r[n],t.indexOf(a)>=0||(s[a]=e[a]);return s}(e,t);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);for(n=0;n<r.length;n++)a=r[n],t.indexOf(a)>=0||Object.prototype.propertyIsEnumerable.call(e,a)&&(s[a]=e[a])}return s}var l=s.a.createContext({}),o=function(e){var t=s.a.useContext(l),a=t;return e&&(a="function"==typeof e?e(t):c(c({},t),e)),a},u=function(e){var t=o(e.components);return s.a.createElement(l.Provider,{value:t},e.children)},p={inlineCode:"code",wrapper:function(e){var t=e.children;return s.a.createElement(s.a.Fragment,{},t)}},d=s.a.forwardRef((function(e,t){var a=e.components,n=e.mdxType,r=e.originalType,i=e.parentName,l=b(e,["components","mdxType","originalType","parentName"]),u=o(a),d=n,m=u["".concat(i,".").concat(d)]||u[d]||p[d]||r;return a?s.a.createElement(m,c(c({ref:t},l),{},{components:a})):s.a.createElement(m,c({ref:t},l))}));function m(e,t){var a=arguments,n=t&&t.mdxType;if("string"==typeof e||n){var r=a.length,i=new Array(r);i[0]=d;var c={};for(var b in t)hasOwnProperty.call(t,b)&&(c[b]=t[b]);c.originalType=e,c.mdxType="string"==typeof e?e:n,i[1]=c;for(var l=2;l<r;l++)i[l]=a[l];return s.a.createElement.apply(null,i)}return s.a.createElement.apply(null,a)}d.displayName="MDXCreateElement"}}]);