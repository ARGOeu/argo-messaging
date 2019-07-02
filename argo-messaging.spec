#debuginfo not supported with Go
%global debug_package %{nil}

Name: argo-messaging
Summary: ARGO Messaging API for broker network
Version: 1.0.4
Release: 1%{?dist}
License: ASL 2.0
Buildroot: %{_tmppath}/%{name}-buildroot
Group: Unspecified
Source0: %{name}-%{version}.tar.gz
BuildRequires: golang
BuildRequires: git
Requires(pre): /usr/sbin/useradd, /usr/bin/getent
ExcludeArch: i386

%description
Installs the ARGO Messaging API

%pre
/usr/bin/getent group argo-messaging || /usr/sbin/groupadd -r argo-messaging
/usr/bin/getent passwd argo-messaging || /usr/sbin/useradd -r -s /sbin/nologin -d /var/www/argo-messaging -g argo-messaging argo-messaging

%prep
%setup

%build
export GOPATH=$PWD
export PATH=$PATH:$GOPATH/bin

cd src/github.com/ARGOeu/argo-messaging/
go get github.com/tools/godep
godep restore
godep update ...
go install

%install
%{__rm} -rf %{buildroot}
install --directory %{buildroot}/var/www/argo-messaging
install --mode 755 bin/argo-messaging %{buildroot}/var/www/argo-messaging/argo-messaging

install --directory %{buildroot}/etc/argo-messaging
install --mode 644 src/github.com/ARGOeu/argo-messaging/config.json %{buildroot}/etc/argo-messaging/config.json

install --directory %{buildroot}/etc/init
install --mode 644 src/github.com/ARGOeu/argo-messaging/argo-messaging.conf %{buildroot}/etc/init/

install --directory %{buildroot}/usr/lib/systemd/system
install --mode 644 src/github.com/ARGOeu/argo-messaging/argo-messaging.service %{buildroot}/usr/lib/systemd/system/

%clean
%{__rm} -rf %{buildroot}
export GOPATH=$PWD
cd src/github.com/ARGOeu/argo-messaging/
go clean

%files
%defattr(0644,argo-messaging,argo-messaging)
%attr(0750,argo-messaging,argo-messaging) /var/www/argo-messaging
%attr(0755,argo-messaging,argo-messaging) /var/www/argo-messaging/argo-messaging
%caps(cap_net_bind_service=+ep) /var/www/argo-messaging/argo-messaging
%config(noreplace) %attr(0644,argo-messaging,argo-messaging) /etc/argo-messaging/config.json
%attr(0644,root,root) /etc/init/argo-messaging.conf
%attr(0644,root,root) /usr/lib/systemd/system/argo-messaging.service

%changelog
* Tue Jul 02 2019 Agelos Tsalapatis  <agelos.tsal@gmail.com> 1.0.4-1%{?dist}
- ARGO-1840 Update the error response for topic:publish and subscription:pull whenever a kafka error is encountered
- Consumer script
- ARGO-1825 Update the request logging format
- AO-492 Make syslog logging configurable for AMS
- ARGO-1801 Update response Verify push endpoint call
- ARGO-1692 Upgrade authorisation per resource handling
- ARGO-1803 Update service file to include service restart on failure
- ARGO-1782 Adjust push worker workflow depending on the verification of the push endpoint of each subscription
- ARGO-1792 API Call - Verify Push Endpoint
- ARGO-1787 Add verification_hash and verified fields for push enabled subscriptions
- ARGO-1683 Block push worker user from pulling when push enabled is false
- ARGO-1723 Republishing of specific messages
- ARGO-1649 API Call that returns a user's profile based on the provided auth token
- ARGO-1721 [GRPC status check] - Update ams push server client to use the new status rpc call
- ARGO-1684 update status call to handle push enabled false
- ARGO-1669 Allow only push worker user to pull from push enabled subscription
- ARGO-1627 Check if the respective topic exists when pulling messages
- ARGO-1632 Add ACL-based access in subscriptions:list
- ARGO-1631 Add ACL-based access to topics:list
- ARGO-1657 Add/remove push worker from sub's acl and link him with sub's projec
- ARGO-1661 Ams handling of push worker initialisation
- ARGO-1656 Internal function - append project to user's projects
- ARGO-1639 API Call - List topic's subscriptions
- ARGO-1651 Internal function - remove user(s) from topic/sub ACL
- ARGO-1650 Internal function - append user(s) to topic/sub ACL
- ARGO-1630 Push worker role
- ARGO-1604 Add health check call for grpc backends
- ARGO-1600 Add push server interaction on modify push config api call
- ARGO-1606 Update push status field api call
- ARGO-1602 Ams push server single connection
- ARGO-1592 ACL for topic/sub should not contain empty names
- ARGO-1553 Grpc client to interafce with the push server
- ARGO-1554 Add a status field at the subscription struct that will contain information regarding its activation on the ams push server
- ARGO-1252 Update config to handle push server information
- ARGO-1550 Disable push functionality in ams
- ARGO-1471 Create a streaming producer
- ARGO-1454 Migrate argo-messaging to golang/dep tool
- ARGO-1469 Create a bulk producer
- ARGO-1446 Improve the receiver endpoint to be more robust
- ARGO-486 Add pagination support for project subscriptions
- ARGO-487 Add pagination support for project topics
- ARGO-1436 Mongo _id field exposure for pagination affects user creation
- ARGO-1432 Add pagination support for users
- ARGO-1431 Add daily msg count for projects:metrics
- ARGO-1399 Topic:metrics && Subscription:metrics check if topic/sub exIsts
- ARGO-1427 Add daily msg count for topics:metrics
- ARGO-1401 Number of messages send via the Argo Messaging Service (per day)
- ARGO-421 Modify sub's ack deadline
- ARGO-1410 Fix nil context bug
- ARGO-1376 Extend ams-migrate script to support import
- ARGO-1375 Script to export AMS kafka data
- ARGO-1373 argo-messaging add failsafe check to not allow admin empty tokens
* Tue Jul 30 2018 Kostas Koumantaros <kkoumantaros@gmail.com> 1.0.3-1%{?dist}
- ARGO-1365 Add config noreplace param in spec file 
- ARGO-1364 Set-cap option in spec file 
- ARGO-1359 Handle empty project_uuid references 
- ARGO-1122 Subscriptions - Set default functionality for pulling messages to return immediately
- ARGO-1279 API CALL - Health check
- ARGO-1307 Update ams service file to include a syslog identifier 
- ARGO-1307 Update ams service file to include a syslog identifier 
- ARGO-1282 Fix Metrics package timestamp to be utc 
- ARGO-1281 Add support for logging to syslog 
- ARGO-571 Use const for error messages in messaging service 
* Tue Jun 12 2018 Kostas Koumantaros <kkoumantaros@gmail.com> 1.0.2-1%{?dist}
- ARGO-1216 Retry if backends are unavailable 
- ARGO-1216 Retry if backends are unavailable 
- ARGO-1177 Fix utc generation in utc-formatted fields 
- ARGO-1177 Fix utc in created,modified fields 
- ARGO-1157 Add get user by Token 
- ARGO-1158 Expose UUID field when querying users 
- ARGO-1154 API CALL - Return User given a UUID 
- ARGO-1085 Add info on Ack timeout error for argo-messaging service 
- ARGO-1003 Fix publishedTime to be in UTC instead of localtime 
* Tue Oct 27 2017 Kostas Koumantaros <kkoumantaros@gmail.com> 1.0.1-1%{?dist}
* Kostas Kaggelidis <kaggis> Added Support for Metrics and CORS
- ARGO-925 Fix return Immediately functionality in pull operation
- ARGO-909 Fix bug on project metrics topics,sub zero values
- ARGO-891 Implement ams request: get User info by Token. Expand user info
- Fix metrics typo. Fix package dependencies
- Add CORS support
- ARGO-859 Add operational metric: memory usage for ams nodes
- ARGO-860 Add CPU Usage metric for ams service nodes
- ARGO-863 Add metric: Aggregation of topics per user at project.
- ARGO-865 aggregation of subscriptions based on project_admin
- Change precedence of project:metrics route
- ARGO-866 Metric: number of subscriptions per topic
- ARGO-862 Add metric: number of topics per project/user
- ARGO-780 Implement Metric: data volume consumed by subscription
- ARGO-779 Implement metric: data volume published to a topic
- ARGO-778 Implement Sub Metric: number of messages consumed
- ARGO-777 Implement metric: number of messages per topic
- ARGO-669 Enable offset changes in subscriptions for event replay
- ARGO-813 Handle gracefully "not found" error during datastore updates
- ARGO-796 Increase consumer default fetch size to handle larger messages
- Updated messaging documentation
- Correct reference to sub/topic in api_subs.md
- Updated example to api_subs documentation  
- ARGO-650 Push endpoint should be https  
- ARGO-646 Sub pull update fix   
- ARGO-640 Add latest topic offset when creating a new subscription  
- ARGO-630 Fix msg id mapping to broker offset issue  
- ARGO-628 Fix offset off bug  
- ARGO-624 Fix consumer acl bug   
- ARGO-615 Add secondary logging of messages that exceed size threshold  
- ARGO-595 Fix users listing null details if user doesn't exist  
- ARGO-519 Implement configurable level-logging  
- ARGO-580 Add command line config parameters and help
* Tue Oct 25 2016 Themis Zamani <themiszamani@gmail.com> - 1.0.0-1%{?dist}
- New RPM package release.
* Thu Mar 24 2016 Themis Zamani <themiszamani@gmail.com> - 0.9.2-1%{?dist}
- ARGO-375 - Added Authentication to Messaging API
- ARGO-324 - Implemented Subscription pull method
- ARGO-323 - Implemented Topic:Publish call
- ARGO-321 - Implemented Topics resource model and calls
- ARGO-320 - Implemented Message Resource
- ARGO-319 - Added initial api frontend
* Thu Jan 21 2016 Konstantinos Kagkelidis <kaggis@gmail.com> - 0.9.1-1%{?dist}
- First Implementation of ARGO API for messaging
- Connect to a Apace Kafka broker network with a list of designated topics
