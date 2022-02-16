(window.webpackJsonp=window.webpackJsonp||[]).push([[7],{62:function(e,t,s){"use strict";s.r(t),s.d(t,"frontMatter",(function(){return r})),s.d(t,"metadata",(function(){return o})),s.d(t,"rightToc",(function(){return c})),s.d(t,"default",(function(){return u}));var a=s(2),i=s(6),n=(s(0),s(91)),r={id:"what_is_ams",title:"AMS - The Service",slug:"/"},o={unversionedId:"what_is_ams",id:"what_is_ams",isDocsHomePage:!1,title:"AMS - The Service",description:"The ARGO Messaging Service (AMS)  is a Publish/Subscribe Service, which implements the Google PubSub protocol. Instead of focusing on a single Messaging API specification for handling the logic of publishing/subscribing to the broker network the API focuses on creating nodes of Publishers and Subscribers as a Service. It provides an HTTP API that enables Users/Systems to implement message oriented service using the Publish/Subscribe Model over plain HTTP.",source:"@site/docs/what_is_ams.md",permalink:"/argo-messaging/docs/",sidebar:"someSidebar",next:{title:"Overview & Introduction",permalink:"/argo-messaging/docs/overview"}},c=[{value:"Features",id:"features",children:[]},{value:"Architectural aspect",id:"architectural-aspect",children:[]},{value:"Fundamentals",id:"fundamentals",children:[{value:"Pull vs Push Subscriptions",id:"pull-vs-push-subscriptions",children:[]}]}],l={rightToc:c};function u(e){var t=e.components,s=Object(i.a)(e,["components"]);return Object(n.b)("wrapper",Object(a.a)({},l,s,{components:t,mdxType:"MDXLayout"}),Object(n.b)("p",null,"The ARGO Messaging Service (AMS)  is a Publish/Subscribe Service, which implements the Google PubSub protocol. Instead of focusing on a single Messaging API specification for handling the logic of publishing/subscribing to the broker network the API focuses on creating nodes of Publishers and Subscribers as a Service. It provides an HTTP API that enables Users/Systems to implement message oriented service using the Publish/Subscribe Model over plain HTTP."),Object(n.b)("p",null,"The ARGO Messaging Service is a real-time messaging service that allows the user to send and receive messages between independent applications. It is implemented as a Publish/Subscribe Service. Instead of focusing on a single Messaging service specification for handling the logic of publishing/subscribing to the broker network the service focuses on creating nodes of Publishers and Subscribers as a Service. In the Publish/Subscribe paradigm, Publishers are users/systems that can send messages to named-channels called Topics. Subscribers are users/systems that create Subscriptions to specific topics and receive messages. "),Object(n.b)("h2",{id:"features"},"Features"),Object(n.b)("ul",null,Object(n.b)("li",{parentName:"ul"},Object(n.b)("strong",{parentName:"li"},"Ease of use"),": It supports an HTTP API and a python library so as to easily integrate with the AMS. "),Object(n.b)("li",{parentName:"ul"},Object(n.b)("strong",{parentName:"li"},"Push Delivery"),": \u0391\u039cS instantly pushes asynchronous event notifications when messages are published to the message topic. Subscribers are notified when a message is available."),Object(n.b)("li",{parentName:"ul"},Object(n.b)("strong",{parentName:"li"},"Replay messages"),": replay messages that have been acknowledged by seeking to a timestamp. "),Object(n.b)("li",{parentName:"ul"},Object(n.b)("strong",{parentName:"li"},"Schema Support"),": on demand mechanism that enables a)  the definition of the expected payload schema, b)  the definition of the expected set of attributes and values and c) the validation for each message if the requirements are met and immediately notify client"),Object(n.b)("li",{parentName:"ul"},Object(n.b)("strong",{parentName:"li"},"Replicate messages on multiple topics"),": Republisher script that consumes and publishes messages for specific topics (ex. SITES) ")),Object(n.b)("p",null,"The AMS supports ",Object(n.b)("strong",{parentName:"p"},"\u201cSchema Validation per topic\u201d"),". It allows the user to define a schema for each topic and validate messages as they are published. It can protect topics from garbage, incomplete messages especially when a topic has multiple remote publishers to ensure data integrity on the client side. "),Object(n.b)("p",null,"The \u201cReplay messages\u201d feature is an offset manipulation mechanism that allows the client on demand to replay or skip messages. When creating a subscription (or editing an existing one), there is an internal option to retain acknowledged messages (by default up to 7 days, or more on request).  To replay and reprocess these messages (ex. testing, error in manipulation etc) , the client has the ability to go back and use the same messages just by seeking a previous timestamp. If  the user needs to skip messages,  he just has to  seek an offset in the future. "),Object(n.b)("p",null,"The implementation for the \u201cpush server\u201d is one of the features used.  The push server(s) are an optional set of worker-machines - deployed on demand - that are needed when the AMS wants to support push enabled subscriptions. The latest implementation provides a gRPC interface in order to communicate with AMS api.  A new security approach was also introduced to enable a secure handshake. "),Object(n.b)("h2",{id:"architectural-aspect"},"Architectural aspect"),Object(n.b)("ul",null,Object(n.b)("li",{parentName:"ul"},Object(n.b)("strong",{parentName:"li"},"Durability"),": provide very high durability, and at-least-once delivery, by storing copies of the same message on multiple servers."),Object(n.b)("li",{parentName:"ul"},Object(n.b)("strong",{parentName:"li"},"Scalability"),": It can handle increases in load without noticeable degradation of latency or availability"),Object(n.b)("li",{parentName:"ul"},Object(n.b)("strong",{parentName:"li"},"Latency"),": A high performance service that can serve more than 1 billion messages per year "),Object(n.b)("li",{parentName:"ul"},Object(n.b)("strong",{parentName:"li"},"Availability"),":  it deals with different types of issues, gracefully failing over in a way that is unnoticeable to end users. Failures can occur in hardware, in software, and due to load.  ")),Object(n.b)("h2",{id:"fundamentals"},"Fundamentals"),Object(n.b)("p",null,"In the Publish/Subscribe paradigm, Publishers are users/systems that can send messages to named-channels called Topics. Subscribers are users/systems that create Subscriptions to specific topics and receive messages."),Object(n.b)("ul",null,Object(n.b)("li",{parentName:"ul"},Object(n.b)("strong",{parentName:"li"},"Topics"),": Topics are resources that can hold messages. Publishers (users/systems) can create topics on demand and name them (Usually with names that make sense and express the class of messages delivered in the topic)"),Object(n.b)("li",{parentName:"ul"},Object(n.b)("strong",{parentName:"li"},"Subscriptions"),": In order for a user to be able to consume messages, he must first create a subscription. Subscriptions are resources that can be created by users on demand and are attached to specific topics. Each topic can have multiple subscriptions but each subscription can be attached to just one topic. Subscriptions allows Subscribers to incrementally consume messages, at their own pace, while the progress is automatically tracked for each subscription."),Object(n.b)("li",{parentName:"ul"},Object(n.b)("strong",{parentName:"li"},"Message"),": The combination of data and (optional) attributes that a publisher sends to a topic and is eventually delivered to subscribers."),Object(n.b)("li",{parentName:"ul"},Object(n.b)("strong",{parentName:"li"},"Message attribute"),": A key-value pair that a publisher can define for a message. ")),Object(n.b)("h3",{id:"pull-vs-push-subscriptions"},"Pull vs Push Subscriptions"),Object(n.b)("p",null,"AMS supports both push and pull message delivery. In push delivery, the Messaging Service initiates requests to your subscriber application to deliver messages. In pull delivery, your subscription application initiates requests to the Pub/Sub server to retrieve messages."),Object(n.b)("h4",{id:"pull-subscriptions"},"Pull subscriptions"),Object(n.b)("p",null,"Pull subscriptions can be configured to require that message deliveries are acknowledged by the Subscribers. If an acknowledgement is made, subscription can resume progressing and send the next available messages. If no acknowledgement is made subscription pauses progressing and re-sends the same messages.\nIn a pull subscription, the subscribing application explicitly calls the API pull method, which requests delivery of a message in the subscription queue. The Pub/Sub server responds with the message (or an error if the queue is empty), and an ack ID. The subscriber then explicitly calls the acknowledge method, using the returned ack ID, to acknowledge receipt."),Object(n.b)("h4",{id:"push-subscriptions"},"Push subscriptions**"),Object(n.b)("p",null,'In a push subscription, the push server sends a request to the subscriber application, at a preconfigured endpoint. The subscriber\'s HTTP response serves as an implicit acknowledgement: a success response indicates that the message has been successfully processed and the Pub/Sub system can delete it from the subscription; a non-success response indicates that the Pub/Sub server should resend it (implicit "nack"). To ensure that subscribers can handle the message flow, the Pub/Sub dynamically adjusts the flow of requests and uses an algorithm to rate-limit retries.\nThe push server(s) are an optional set of worker-machines that are needed when the AMS wants to support push enabled subscriptions.\nIt allows to decouple the push functionality from AMS api nodes\nThey perform the push functionality for the messages of a push enabled subscription (consume->deliver\u2192ack)/\nProvide a gRPC interface in order to communicate with their api\nProvide subscription runtime status'),Object(n.b)("p",null,Object(n.b)("strong",{parentName:"p"},"Apart from all these the Messaging Service supports:")),Object(n.b)("ul",null,Object(n.b)("li",{parentName:"ul"},Object(n.b)("strong",{parentName:"li"},"Argo-ams-library"),": A simple library to interact with the ARGO Messaging Service."),Object(n.b)("li",{parentName:"ul"},Object(n.b)("strong",{parentName:"li"},"Argo-AuthN"),": Argo-authn is a new Authentication Service. This service provides the ability to different services to use alternative authentication mechanisms without having to store additional user info or implement new functionalities.The AUTH service holds various information about a service\u2019s users, hosts, API urls, etc, and leverages them to provide its functionality."),Object(n.b)("li",{parentName:"ul"},Object(n.b)("strong",{parentName:"li"},"AMS Metrics"),": Metrics about the service and the usage.")))}u.isMDXComponent=!0},91:function(e,t,s){"use strict";s.d(t,"a",(function(){return p})),s.d(t,"b",(function(){return d}));var a=s(0),i=s.n(a);function n(e,t,s){return t in e?Object.defineProperty(e,t,{value:s,enumerable:!0,configurable:!0,writable:!0}):e[t]=s,e}function r(e,t){var s=Object.keys(e);if(Object.getOwnPropertySymbols){var a=Object.getOwnPropertySymbols(e);t&&(a=a.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),s.push.apply(s,a)}return s}function o(e){for(var t=1;t<arguments.length;t++){var s=null!=arguments[t]?arguments[t]:{};t%2?r(Object(s),!0).forEach((function(t){n(e,t,s[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(s)):r(Object(s)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(s,t))}))}return e}function c(e,t){if(null==e)return{};var s,a,i=function(e,t){if(null==e)return{};var s,a,i={},n=Object.keys(e);for(a=0;a<n.length;a++)s=n[a],t.indexOf(s)>=0||(i[s]=e[s]);return i}(e,t);if(Object.getOwnPropertySymbols){var n=Object.getOwnPropertySymbols(e);for(a=0;a<n.length;a++)s=n[a],t.indexOf(s)>=0||Object.prototype.propertyIsEnumerable.call(e,s)&&(i[s]=e[s])}return i}var l=i.a.createContext({}),u=function(e){var t=i.a.useContext(l),s=t;return e&&(s="function"==typeof e?e(t):o(o({},t),e)),s},p=function(e){var t=u(e.components);return i.a.createElement(l.Provider,{value:t},e.children)},b={inlineCode:"code",wrapper:function(e){var t=e.children;return i.a.createElement(i.a.Fragment,{},t)}},h=i.a.forwardRef((function(e,t){var s=e.components,a=e.mdxType,n=e.originalType,r=e.parentName,l=c(e,["components","mdxType","originalType","parentName"]),p=u(s),h=a,d=p["".concat(r,".").concat(h)]||p[h]||b[h]||n;return s?i.a.createElement(d,o(o({ref:t},l),{},{components:s})):i.a.createElement(d,o({ref:t},l))}));function d(e,t){var s=arguments,a=t&&t.mdxType;if("string"==typeof e||a){var n=s.length,r=new Array(n);r[0]=h;var o={};for(var c in t)hasOwnProperty.call(t,c)&&(o[c]=t[c]);o.originalType=e,o.mdxType="string"==typeof e?e:a,r[1]=o;for(var l=2;l<n;l++)r[l]=s[l];return i.a.createElement.apply(null,r)}return i.a.createElement.apply(null,s)}h.displayName="MDXCreateElement"}}]);