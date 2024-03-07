"use strict";(self.webpackChunkwebsite=self.webpackChunkwebsite||[]).push([[1085],{6802:(e,s,t)=>{t.r(s),t.d(s,{assets:()=>c,contentTitle:()=>o,default:()=>l,frontMatter:()=>i,metadata:()=>r,toc:()=>d});var n=t(4848),a=t(8453);const i={id:"publisher",title:"Publisher Guide",sidebar_position:1},o=void 0,r={id:"guides/publisher",title:"Publisher Guide",description:"Publishers can send messages to named-channels called Topics.",source:"@site/docs/guides/publisher.md",sourceDirName:"guides",slug:"/guides/publisher",permalink:"/argo-messaging/docs/guides/publisher",draft:!1,unlisted:!1,tags:[],version:"current",sidebarPosition:1,frontMatter:{id:"publisher",title:"Publisher Guide",sidebar_position:1},sidebar:"tutorialSidebar",previous:{title:"Guides",permalink:"/argo-messaging/docs/category/guides"},next:{title:"Subscriber Guide",permalink:"/argo-messaging/docs/guides/subscriber_guide"}},c={},d=[{value:"Before you start",id:"before-you-start",level:2},{value:"Start publishing",id:"start-publishing",level:2},{value:"Schema Support",id:"schema-support",level:2}];function h(e){const s={a:"a",code:"code",h2:"h2",li:"li",p:"p",pre:"pre",strong:"strong",ul:"ul",...(0,a.R)(),...e.components};return(0,n.jsxs)(n.Fragment,{children:[(0,n.jsx)(s.p,{children:"Publishers can send messages to named-channels called Topics."}),"\n",(0,n.jsx)(s.h2,{id:"before-you-start",children:"Before you start"}),"\n",(0,n.jsxs)(s.p,{children:["In order to get an account on the ARGO Messaging Service, submit a request through the ",(0,n.jsx)(s.a,{href:"https://ams-register.argo.grnet.gr/",children:"ARGO Messaging Service account form"})]}),"\n",(0,n.jsx)(s.p,{children:"Upon account approval, you will receive information via e-mail about your new project along with an API token."}),"\n",(0,n.jsx)(s.h2,{id:"start-publishing",children:"Start publishing"}),"\n",(0,n.jsx)(s.p,{children:"When everything is set up you can start by following the general flow for a publisher:"}),"\n",(0,n.jsxs)(s.p,{children:[(0,n.jsx)(s.strong,{children:"Step 1:"})," Create a topic"]}),"\n",(0,n.jsxs)(s.p,{children:["For more details visit section ",(0,n.jsx)(s.a,{href:"/argo-messaging/docs/api_advanced/api_topics#create-topic",children:"Topics: Create a topic"})]}),"\n",(0,n.jsxs)(s.p,{children:[(0,n.jsx)(s.strong,{children:"Step 2:"})," Create a subscription"]}),"\n",(0,n.jsxs)(s.p,{children:["A Topic without at least one Subscription act like black holes. Publishers can send messages to those topics, but the messages will not be retrievable. In order to be able to publish and consume messages, at least one Subscription must created to the Topic that you are publishing messages to. By default, a Subscription is created in pull mode, meaning that consumers can query the Messaging API and retrieve the messages that are published to the Topic that the Subscription is configured for. More information about how create a Subscription, visit section ",(0,n.jsx)(s.a,{href:"/argo-messaging/docs/api_advanced/api_subscriptions#create-subs",children:"Subscriptions: Create a subscription"})]}),"\n",(0,n.jsxs)(s.p,{children:[(0,n.jsx)(s.strong,{children:"Step 3:"})," Start publishing messages"]}),"\n",(0,n.jsx)(s.p,{children:"The ARGO Messaging Service accepts JSON over HTTP. In order to publish messages you have to represent them using the following schema:"}),"\n",(0,n.jsx)(s.pre,{children:(0,n.jsx)(s.code,{className:"language-json",children:'{\n  "$schema": "http://json-schema.org/draft-04/schema#",\n  "type": "object",\n  "properties": {\n    "messages": {\n      "type": "array",\n      "items": {\n        "type": "object",\n        "anyOf": [{\n          "properties": {\n            "data": {\n              "type": "string",\n              "contentEncoding": "base64",\n              "minLength": 1\n            },\n          },\n          "required": ["data"]\n        },{\n          "properties": {\n            "attributes": {\n              "type": "object",\n              "minProperties": 1,\n              "properties": {}\n            }\n          },\n          "required": ["attributes"]\n        }]\n      }\n    }\n  },\n  "required": [\n    "messages"\n  ]\n}\n'})}),"\n",(0,n.jsx)(s.p,{children:"The JSON body send to the ARGO Messaging Service may contain one or more messages. Each message can have:"}),"\n",(0,n.jsxs)(s.ul,{children:["\n",(0,n.jsx)(s.li,{children:"attributes: optional key value pair of metadata you desire"}),"\n",(0,n.jsx)(s.li,{children:"data: the data of the message."}),"\n"]}),"\n",(0,n.jsx)(s.p,{children:"The data must be base64-encoded, and can not exceed 10MB after encoding. Note that the message payload must not be empty; it must contain either a non-empty data field, or at least one attribute."}),"\n",(0,n.jsx)(s.p,{children:"Below you can find an example, in which a user publishes two messages in one call:"}),"\n",(0,n.jsx)(s.pre,{children:(0,n.jsx)(s.code,{className:"language-json",children:'{\n  "messages": [\n  {\n    "attributes":\n    {\n      "station":"NW32ZC",\n      "status":"PROD"\n    },\n    "data":"U28geW91IHdlbnQgYWhlYWQgYW5kIGRlY29kZWQgdGhpcywgeW91IGNvdWxkbid0IHJlc2lzdCBlaCA/"\n  },\n  {\n    "attributes":\n    {\n      "station":"GHJ32",\n      "status":"TEST"\n    },\n    "data":"U28geW91IHdlbnQgYWhlYWQgYW5kIGRlY29kZWQgdGhpcywgeW91IGNvdWxkbid0IHJlc2lzdCBlaCA/"\n  }\n  ]\n}\n'})}),"\n",(0,n.jsx)(s.p,{children:"You can publish and consume any kind of data through the ARGO Messaging Service (as long as the base64 encoded payload is not larger than the maximum acceptable size)."}),"\n",(0,n.jsxs)(s.p,{children:["For more details visit section ",(0,n.jsx)(s.a,{href:"/argo-messaging/docs/api_advanced/api_topics#publish",children:"Topics: Publish message/s to a topic"})]}),"\n",(0,n.jsx)(s.h2,{id:"schema-support",children:"Schema Support"}),"\n",(0,n.jsx)(s.p,{children:"The AMS supports \u201cSchema Validation per topic\u201d."}),"\n",(0,n.jsx)(s.p,{children:"When a user want to support a predefined format for messages exchanged then a topic should be created with a schema defined for it.\nA schema is a format that messages from a topic must follow. It is actually a contract between publisher and subscriber. The Messaging Service allows the user to define a schema for each topic and validate messages as they are published. It can protect topics from garbage, incomplete messages especially when a topic has multiple remote publishers to ensure data integrity on the client side."}),"\n",(0,n.jsx)(s.p,{children:"The Schema Support is on demand mechanism that enables a) the definition of the expected payload schema, b) the definition of the expected set of attributes and values and c) the validation for each message if the requirements are met and immediately notify client"}),"\n",(0,n.jsx)(s.p,{children:"The steps that you should follow for a schema support"}),"\n",(0,n.jsxs)(s.p,{children:[(0,n.jsx)(s.strong,{children:"Step 1:"})," Create a new schema in your project"]}),"\n",(0,n.jsx)(s.p,{children:"The Supported Schema Types are JSON and AVRO"}),"\n",(0,n.jsxs)(s.p,{children:["For more details visit section  ",(0,n.jsx)(s.a,{href:"/argo-messaging/docs/api_advanced/api_schemas#create-schema",children:"Create new schema"})]}),"\n",(0,n.jsxs)(s.p,{children:[(0,n.jsx)(s.strong,{children:"Step 2:"})," Create a topic with this schema attached"]}),"\n",(0,n.jsx)(s.p,{children:"If you need to link a schema with your topic you need to provide its name, to the api call during the creation of the topic.."}),"\n",(0,n.jsxs)(s.p,{children:["For more details visit section ",(0,n.jsx)(s.a,{href:"/argo-messaging/docs/api_advanced/api_topics#create-topic",children:"Create new topic"})]}),"\n",(0,n.jsxs)(s.p,{children:[(0,n.jsx)(s.strong,{children:"Step 3:"})," Assign this schema to your topic"]}),"\n",(0,n.jsx)(s.p,{children:"If you need to link a schema with your topic you need to provide its name, to the api call"}),"\n",(0,n.jsxs)(s.p,{children:["For more details visit section ",(0,n.jsx)(s.a,{href:"/argo-messaging/docs/api_advanced/api_topics#create-topic",children:"Update the topic"})]}),"\n",(0,n.jsxs)(s.p,{children:[(0,n.jsx)(s.strong,{children:"Step 4:"})," Validate the message"]}),"\n",(0,n.jsx)(s.p,{children:"This  is used whenever we want to test a message against a schema. The process to check that your schema and messages are working as expected is to create a new topic that needs to be associated with the schema, then create the message in base64 encoding and publish it to the topic. Instead of creating all this pipeline in order to check your schema and messages we can explicitly do it on this API call."}),"\n",(0,n.jsxs)(s.p,{children:["For more details visit section ",(0,n.jsx)(s.a,{href:"/argo-messaging/docs/api_advanced/api_schemas#validate",children:"Validate the message"})]}),"\n",(0,n.jsxs)(s.p,{children:[(0,n.jsx)(s.strong,{children:"Step 5:"})," Publish messages to your topic"]}),"\n",(0,n.jsx)(s.p,{children:"You may now start publishing messages to your topic."}),"\n",(0,n.jsxs)(s.p,{children:["For more details visit section ",(0,n.jsx)(s.a,{href:"/argo-messaging/docs/api_advanced/api_topics#publish",children:"publish-messages-to-a-topic"})]})]})}function l(e={}){const{wrapper:s}={...(0,a.R)(),...e.components};return s?(0,n.jsx)(s,{...e,children:(0,n.jsx)(h,{...e})}):h(e)}},8453:(e,s,t)=>{t.d(s,{R:()=>o,x:()=>r});var n=t(6540);const a={},i=n.createContext(a);function o(e){const s=n.useContext(i);return n.useMemo((function(){return"function"==typeof e?e(s):{...s,...e}}),[s,e])}function r(e){let s;return s=e.disableParentContext?"function"==typeof e.components?e.components(a):e.components||a:o(e.components),n.createElement(i.Provider,{value:s},e.children)}}}]);