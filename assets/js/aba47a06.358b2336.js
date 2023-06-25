"use strict";(self.webpackChunkwebsite=self.webpackChunkwebsite||[]).push([[3416],{3905:(e,t,n)=>{n.d(t,{Zo:()=>m,kt:()=>h});var r=n(7294);function s(e,t,n){return t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function a(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),n.push.apply(n,r)}return n}function i(e){for(var t=1;t<arguments.length;t++){var n=null!=arguments[t]?arguments[t]:{};t%2?a(Object(n),!0).forEach((function(t){s(e,t,n[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):a(Object(n)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(n,t))}))}return e}function o(e,t){if(null==e)return{};var n,r,s=function(e,t){if(null==e)return{};var n,r,s={},a=Object.keys(e);for(r=0;r<a.length;r++)n=a[r],t.indexOf(n)>=0||(s[n]=e[n]);return s}(e,t);if(Object.getOwnPropertySymbols){var a=Object.getOwnPropertySymbols(e);for(r=0;r<a.length;r++)n=a[r],t.indexOf(n)>=0||Object.prototype.propertyIsEnumerable.call(e,n)&&(s[n]=e[n])}return s}var u=r.createContext({}),l=function(e){var t=r.useContext(u),n=t;return e&&(n="function"==typeof e?e(t):i(i({},t),e)),n},m=function(e){var t=l(e.components);return r.createElement(u.Provider,{value:t},e.children)},p={inlineCode:"code",wrapper:function(e){var t=e.children;return r.createElement(r.Fragment,{},t)}},c=r.forwardRef((function(e,t){var n=e.components,s=e.mdxType,a=e.originalType,u=e.parentName,m=o(e,["components","mdxType","originalType","parentName"]),c=l(n),h=s,d=c["".concat(u,".").concat(h)]||c[h]||p[h]||a;return n?r.createElement(d,i(i({ref:t},m),{},{components:n})):r.createElement(d,i({ref:t},m))}));function h(e,t){var n=arguments,s=t&&t.mdxType;if("string"==typeof e||s){var a=n.length,i=new Array(a);i[0]=c;var o={};for(var u in t)hasOwnProperty.call(t,u)&&(o[u]=t[u]);o.originalType=e,o.mdxType="string"==typeof e?e:s,i[1]=o;for(var l=2;l<a;l++)i[l]=n[l];return r.createElement.apply(null,i)}return r.createElement.apply(null,n)}c.displayName="MDXCreateElement"},5063:(e,t,n)=>{n.r(t),n.d(t,{assets:()=>u,contentTitle:()=>i,default:()=>p,frontMatter:()=>a,metadata:()=>o,toc:()=>l});var r=n(7462),s=(n(7294),n(3905));const a={id:"qa",title:"Q & A - General Questions",sidebar_position:1},i=void 0,o={unversionedId:"faq/qa",id:"faq/qa",title:"Q & A - General Questions",description:"Questions and answers based on problems encountered during implementation.",source:"@site/docs/faq/qa.md",sourceDirName:"faq",slug:"/faq/qa",permalink:"/argo-messaging/docs/faq/qa",draft:!1,tags:[],version:"current",sidebarPosition:1,frontMatter:{id:"qa",title:"Q & A - General Questions",sidebar_position:1},sidebar:"tutorialSidebar",previous:{title:"Frequent Questions",permalink:"/argo-messaging/docs/category/frequent-questions"},next:{title:"Q & A - Ruby specifics",permalink:"/argo-messaging/docs/faq/qa_ruby"}},u={},l=[{value:"When I pull down messages, they have an publishTime associated with them. What time zone is this field in?",id:"when-i-pull-down-messages-they-have-an-publishtime-associated-with-them-what-time-zone-is-this-field-in",level:2},{value:"if there aren\u2019t enough messages to supply the requested max_messages number, the request eventually returns however many messages are present. Is there a way to configure this timeout, in the post body for example?",id:"if-there-arent-enough-messages-to-supply-the-requested-max_messages-number-the-request-eventually-returns-however-many-messages-are-present-is-there-a-way-to-configure-this-timeout-in-the-post-body-for-example",level:2}],m={toc:l};function p(e){let{components:t,...n}=e;return(0,s.kt)("wrapper",(0,r.Z)({},m,n,{components:t,mdxType:"MDXLayout"}),(0,s.kt)("p",null,"Questions and answers based on problems encountered during implementation. "),(0,s.kt)("h2",{id:"when-i-pull-down-messages-they-have-an-publishtime-associated-with-them-what-time-zone-is-this-field-in"},"When I pull down messages, they have an publishTime associated with them. What time zone is this field in?"),(0,s.kt)("p",null,"The publishTime is the Timestamp associated with the message when the message was published, in UTC Zulu time format - detailed to nanoseconds. - GENERATED BY THE API (UTC+2 at devel infrastructure)"),(0,s.kt)("h2",{id:"if-there-arent-enough-messages-to-supply-the-requested-max_messages-number-the-request-eventually-returns-however-many-messages-are-present-is-there-a-way-to-configure-this-timeout-in-the-post-body-for-example"},"if there aren\u2019t enough messages to supply the requested max_messages number, the request eventually returns however many messages are present. Is there a way to configure this timeout, in the post body for example?"),(0,s.kt)("p",null,"The current timeout is set to 5mins. But the user - client cannot change it. You can optionally set the ",(0,s.kt)("strong",{parentName:"p"},"returnImmediately")," field to ",(0,s.kt)("strong",{parentName:"p"},"true")," to prevent the subscriber from waiting if the queue is currently empty."))}p.isMDXComponent=!0}}]);