#debuginfo not supported with Go
%global debug_package %{nil}

Name: argo-messaging
Summary: ARGO Messaging API for broker network
Version: 1.0.3
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
