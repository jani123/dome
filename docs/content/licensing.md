---
title: Licensing and customer ownership
weight: 80
---

Dome is licensed under **GPL-3.0-only, without a linking exception**. Paid
implementation, installation, and support are welcome. The purpose of this
choice is for customers to retain the ability to maintain their own agents.

Distributing a combined Go agent linked with Dome generally requires providing
that agent under GPLv3 and making its complete corresponding source available,
including the build scripts needed to modify and rebuild it. Providing only
Dome's source is insufficient for that combined work. Private use does not
require publishing source, and recipients need not publish it publicly merely
because they receive it. They retain GPL redistribution/modification rights.

Merely managing a separate application with Dome does not automatically place
that application under GPLv3. The source requirements concern the covered agent.
See the [GPL library FAQ](https://www.gnu.org/licenses/gpl-faq.html#IfLibraryIsGPL)
and [GPLv3 sections 1 and 6](https://www.gnu.org/licenses/gpl-3.0.html).

Recommended customer handover includes the exact agent source, `go.mod`/`go.sum`,
build instructions, applicable license/notices, and infrastructure configuration
alongside the binary. GPLv3 permits multiple compliant source-delivery methods;
this handover recommendation is not an additional license restriction.

Give the customer control of their source/configuration repositories, deployment
credentials, and SOPS keys. Transfer secrets through appropriate private channels,
not public source archives. Source availability alone cannot guarantee access to
machines or credentials. The customer should be able to change providers.
