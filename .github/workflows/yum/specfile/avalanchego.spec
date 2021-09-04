%define _build_id_links none

Name:           paaro
Version:        %{version}
Release:        %{release}
Summary:        The Dijets platform binaries
URL:            https://github.com/djt-labs/%{name}
License:        BSD-3
AutoReqProv:    no

%description
Dijets is an incredibly lightweight protocol, so the minimum computer requirements are quite modest.

%files
/usr/local/bin/paaro
/usr/local/lib/paaro
/usr/local/lib/paaro/evm

%changelog
* Mon Oct 26 2020 Charlie Wyse <jack@djtlabs.dijets.io>
- First creation of package

