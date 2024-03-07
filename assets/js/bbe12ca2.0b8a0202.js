"use strict";(self.webpackChunkwebsite=self.webpackChunkwebsite||[]).push([[9817],{468:(e,n,s)=>{s.r(n),s.d(n,{assets:()=>r,contentTitle:()=>o,default:()=>g,frontMatter:()=>i,metadata:()=>c,toc:()=>d});var t=s(4848),a=s(8453);const i={id:"msg_backend",title:"Using Apache Kafka as a Backend Message system",sidebar_position:2},o=void 0,c={id:"concepts/msg_backend",title:"Using Apache Kafka as a Backend Message system",description:"The ARGO Messaging API has been designed to rely on a generic Message Back-end Interface and use specific implementation of that interface for supporting different systems. Right now the first implementation for the messaging back-end relies on Apache Kafka as a distributed messaging system.",source:"@site/docs/concepts/msg_backend.md",sourceDirName:"concepts",slug:"/concepts/msg_backend",permalink:"/argo-messaging/docs/concepts/msg_backend",draft:!1,unlisted:!1,tags:[],version:"current",sidebarPosition:2,frontMatter:{id:"msg_backend",title:"Using Apache Kafka as a Backend Message system",sidebar_position:2},sidebar:"tutorialSidebar",previous:{title:"Overview & Introduction",permalink:"/argo-messaging/docs/concepts/overview"},next:{title:"Data flow in Argo Messaging",permalink:"/argo-messaging/docs/concepts/msg_flow"}},r={},d=[];function p(e){const n={p:"p",...(0,a.R)(),...e.components};return(0,t.jsxs)(t.Fragment,{children:[(0,t.jsx)(n.p,{children:"The ARGO Messaging API has been designed to rely on a generic Message Back-end Interface and use specific implementation of that interface for supporting different systems. Right now the first implementation for the messaging back-end relies on Apache Kafka as a distributed messaging system."}),"\n",(0,t.jsx)(n.p,{children:"A big advantage of the ARGO Messaging API is that provides a mechanism to easily support namespacing and different tenants on a Kafka Back-end (Apache Kafka doesn\u2019t support natively namespacing yet). ARGO Messaging API uses the notion of \u2018projects\u2019 for each tenant and can support multiple projects each one containing multiple topics/subscriptions and users on the same Kafka back-end."})]})}function g(e={}){const{wrapper:n}={...(0,a.R)(),...e.components};return n?(0,t.jsx)(n,{...e,children:(0,t.jsx)(p,{...e})}):p(e)}},8453:(e,n,s)=>{s.d(n,{R:()=>o,x:()=>c});var t=s(6540);const a={},i=t.createContext(a);function o(e){const n=t.useContext(i);return t.useMemo((function(){return"function"==typeof e?e(n):{...n,...e}}),[n,e])}function c(e){let n;return n=e.disableParentContext?"function"==typeof e.components?e.components(a):e.components||a:o(e.components),t.createElement(i.Provider,{value:n},e.children)}}}]);