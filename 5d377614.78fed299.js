(window.webpackJsonp=window.webpackJsonp||[]).push([[11],{121:function(e,t,n){"use strict";n.r(t),t.default=n.p+"assets/images/flow1_3-98cbe0993d16937912c3737f086d0e12.png"},122:function(e,t,n){"use strict";n.r(t),t.default=n.p+"assets/images/flow4-84b045f31868acc4ed095a5023e68472.png"},123:function(e,t,n){"use strict";n.r(t),t.default=n.p+"assets/images/flow5_6-5e5ad69ed3c2a3a1515d405cba3d99a5.png"},124:function(e,t,n){"use strict";n.r(t),t.default=n.p+"assets/images/multisub-76a59ff3b697cd763db35c24e0946660.png"},66:function(e,t,n){"use strict";n.r(t),n.d(t,"frontMatter",(function(){return o})),n.d(t,"metadata",(function(){return s})),n.d(t,"rightToc",(function(){return c})),n.d(t,"default",(function(){return u}));var r=n(2),a=n(6),i=(n(0),n(90)),o={id:"msg_flow",title:"Data flow in Argo Messaging"},s={unversionedId:"msg_flow",id:"msg_flow",isDocsHomePage:!1,title:"Data flow in Argo Messaging",description:"The main steps of the messaging API:",source:"@site/docs/msg_flow.md",permalink:"/argo-messaging/docs/msg_flow",sidebar:"someSidebar",previous:{title:"Using Apache Kafka as a Backend Message system",permalink:"/argo-messaging/docs/msg_backend"},next:{title:"Authentication & Authorization",permalink:"/argo-messaging/docs/auth"}},c=[],l={rightToc:c};function u(e){var t=e.components,o=Object(a.a)(e,["components"]);return Object(i.b)("wrapper",Object(r.a)({},l,o,{components:t,mdxType:"MDXLayout"}),Object(i.b)("p",null,"The main steps of the messaging API:"),Object(i.b)("ul",null,Object(i.b)("li",{parentName:"ul"},"A user creates a Topic"),Object(i.b)("li",{parentName:"ul"},"Users that want to consume a message set up subscriptions."),Object(i.b)("li",{parentName:"ul"},"Each subscription is set on one Topic"),Object(i.b)("li",{parentName:"ul"},"A Topic can have multiple Subscriptions"),Object(i.b)("li",{parentName:"ul"},"Each subscription sets up a sync point in time."),Object(i.b)("li",{parentName:"ul"},"Messages that are published after that sync point can be pull by or push to the subscribers."),Object(i.b)("li",{parentName:"ul"},"Messages that have been published to the Topic that the Subscription was configured for  before the creation of the Subscription, will not be delivered to the Subscribers."),Object(i.b)("li",{parentName:"ul"},"Each Topic has a TTL values for each messages published to it. Older messages are purged."),Object(i.b)("li",{parentName:"ul"},"Message deliveries can be out-of-order and might have duplicate messages. Each Subscriber should be idempotent"),Object(i.b)("li",{parentName:"ul"},"A Subscription is configured either as in PULL or on PUSH mode. PUSH mode receives a client URI in order to POST messages there")),Object(i.b)("p",null,Object(i.b)("img",{alt:"Flow: Steps 1 to 3 ",src:n(121).default})),Object(i.b)("p",null,Object(i.b)("img",{alt:"Flow: Step 4 ",src:n(122).default})),Object(i.b)("p",null,Object(i.b)("img",{alt:"Flow: Steps 5 to 6 ",src:n(123).default})),Object(i.b)("p",null,"A Topic might have multiple subscriptions and each subscription has it\u2019s own tracked offset on the topic."),Object(i.b)("p",null,Object(i.b)("img",{alt:"Multiple Subscriptions ",src:n(124).default})),Object(i.b)("ul",null,Object(i.b)("li",{parentName:"ul"},"Above: A single Topic holding multiple Subscriptions")))}u.isMDXComponent=!0},90:function(e,t,n){"use strict";n.d(t,"a",(function(){return p})),n.d(t,"b",(function(){return m}));var r=n(0),a=n.n(r);function i(e,t,n){return t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function o(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),n.push.apply(n,r)}return n}function s(e){for(var t=1;t<arguments.length;t++){var n=null!=arguments[t]?arguments[t]:{};t%2?o(Object(n),!0).forEach((function(t){i(e,t,n[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):o(Object(n)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(n,t))}))}return e}function c(e,t){if(null==e)return{};var n,r,a=function(e,t){if(null==e)return{};var n,r,a={},i=Object.keys(e);for(r=0;r<i.length;r++)n=i[r],t.indexOf(n)>=0||(a[n]=e[n]);return a}(e,t);if(Object.getOwnPropertySymbols){var i=Object.getOwnPropertySymbols(e);for(r=0;r<i.length;r++)n=i[r],t.indexOf(n)>=0||Object.prototype.propertyIsEnumerable.call(e,n)&&(a[n]=e[n])}return a}var l=a.a.createContext({}),u=function(e){var t=a.a.useContext(l),n=t;return e&&(n="function"==typeof e?e(t):s(s({},t),e)),n},p=function(e){var t=u(e.components);return a.a.createElement(l.Provider,{value:t},e.children)},b={inlineCode:"code",wrapper:function(e){var t=e.children;return a.a.createElement(a.a.Fragment,{},t)}},f=a.a.forwardRef((function(e,t){var n=e.components,r=e.mdxType,i=e.originalType,o=e.parentName,l=c(e,["components","mdxType","originalType","parentName"]),p=u(n),f=r,m=p["".concat(o,".").concat(f)]||p[f]||b[f]||i;return n?a.a.createElement(m,s(s({ref:t},l),{},{components:n})):a.a.createElement(m,s({ref:t},l))}));function m(e,t){var n=arguments,r=t&&t.mdxType;if("string"==typeof e||r){var i=n.length,o=new Array(i);o[0]=f;var s={};for(var c in t)hasOwnProperty.call(t,c)&&(s[c]=t[c]);s.originalType=e,s.mdxType="string"==typeof e?e:r,o[1]=s;for(var l=2;l<i;l++)o[l]=n[l];return a.a.createElement.apply(null,o)}return a.a.createElement.apply(null,n)}f.displayName="MDXCreateElement"}}]);