# Gotham Rescue Network — Functional Requirements

Updated: 2026-07-31

## Product brief

Gotham Rescue Network (GRN) helps the Gotham Emergency Operations Center coordinate incidents and field units during ordinary nights and city-wide crises.

An operator receives a report, verifies it, requests an appropriate response, observes the response in real time, and closes the incident with a complete history. Field units report availability and location, receive offers, accept or reject them, travel to the scene, and report the outcome. Supervisors intervene when automation or normal procedure is insufficient.

The initial product is for emergency coordination, not public social networking, predictive policing, or autonomous use of force. It may recommend and coordinate responders, but a human remains accountable for exceptional decisions.

This document defines functional behavior. It intentionally does not prescribe service boundaries, storage engines, message brokers, deployment topology, or internal API design.

## Product users

### Gotham resident

Reports an incident and follows the limited public status of that report. A resident must not see responder identities, exact unit locations, operational notes, or other confidential information.

### Emergency operator

Validates reports, sets priority and required capabilities, starts dispatch, watches active incidents, communicates with units, and closes routine incidents.

### Field unit

A police, fire, medical, infrastructure, or special-response team. A unit reports its availability and location, receives assignment offers, and updates its response status.

### Batman / independent special responder

Receives narrowly scoped requests for exceptional incidents through an authorized liaison channel. Batman does not expose a civilian identity, permanent base, or continuously visible location. The operations center can request assistance and receive explicit acknowledgements and operational updates, but cannot command Batman as an ordinary city unit.

### Shift supervisor

Oversees city operations, changes priorities, overrides assignments, approves exceptional actions, and reviews response quality.

### System administrator

Manages users, unit records, operational zones, incident categories, and other controlled configuration. An administrator cannot silently alter incident history.

### Auditor

Investigates what happened after an incident. The auditor can view decisions and changes but cannot operate active incidents.

## Shared terminology

- **Incident:** a reported situation requiring assessment or response.
- **Report:** information received from a resident, sensor, operator, or partner agency. Several reports may refer to one incident.
- **Unit:** a deployable response team or vehicle.
- **Capability:** something a unit can provide, such as `medical`, `fire_suppression`, `bomb_disposal`, `high_angle_rescue`, or `armored_response`.
- **Offer:** a time-limited request asking a unit to accept an assignment.
- **Assignment:** the confirmed relationship between an incident and a responding unit.
- **Operational zone:** a named area of Gotham used for responsibility and reporting.
- **Timeline event:** an immutable record of a meaningful action or state change.
- **Sensitive incident:** an incident whose details are restricted to specifically authorized personnel.

## Global functional rules

1. Every incident has a public reference, an internal identifier, a current status, a priority, a location, an owning operator, and a chronological timeline.
2. A user sees and performs only actions allowed by their role and operational scope.
3. Every consequential action records who performed it, when it occurred, the previous value, the new value, and an optional reason.
4. Repeating the same submitted action because of a retry must not create a second incident, offer, assignment, or status transition.
5. Conflicting actions must produce an explicit result. The system must not silently choose a winner without recording the conflict.
6. User-facing times use Gotham local time, while records must remain unambiguous across time zones and daylight-saving changes.
7. Free-text operational notes are internal unless explicitly marked safe for residents.
8. Historical events cannot be edited or deleted through ordinary product actions. Corrections are appended as new events.
9. The system distinguishes “unknown” from a negative claim. For example, an unknown unit location is not the same as a confirmed location outside an area.
10. Every stage must preserve behavior delivered by previous stages unless a documented product change replaces it.

---

## Stage 1 — Incident desk and manual response

### Goal

Give one emergency operator enough functionality to register a report, create an incident, manually assign a known unit, follow its progress, and close the case with a trustworthy timeline.

### 1.1 Register a resident report

**Used by:** emergency operator.

**Inputs:**

- Source: phone call, walk-in, partner agency, or unknown.
- Reporter name and callback number, both optional.
- Incident category: medical emergency, fire, violent crime, infrastructure failure, suspicious activity, traffic incident, or other.
- Description in the reporter's own words.
- Location as a street address, landmark, or map coordinates.
- Whether danger is currently in progress.
- Whether people are injured or trapped: yes, no, or unknown.
- Client-generated submission identifier.

**Outputs:**

- Public report reference suitable for sharing with the resident.
- Internal incident identifier.
- Normalized location and operational zone, or a clear request for operator correction.
- Initial incident status and priority.
- Recorded creation time and owning operator.

**Rules:**

- Category, description, and some usable form of location are required.
- The same submission identifier returns the previously created result instead of creating a duplicate.
- The operator may correct an unrecognized address before completing registration.
- Reporter contact information is never included in the resident-visible response after initial confirmation.
- The initial priority is suggested from the answers, but the operator confirms it.

### 1.2 View the active incident board

**Used by:** emergency operator and shift supervisor.

**Inputs:**

- Optional filters: status, priority, category, zone, owning operator, assigned unit, and creation-time range.
- Sorting choice: highest priority, oldest unassigned, newest, or last updated.
- Page/cursor for the next result set.

**Outputs:**

- A bounded list of incident summaries.
- For each incident: reference, category, priority, approximate location, age, status, assigned units, owning operator, and time since last meaningful update.
- Counts of active incidents grouped by priority and status.
- A continuation value when more results exist.

**Rules:**

- Sensitive details are omitted from list views unless the user has permission.
- Two incidents with equal priority are ordered predictably.
- A newly created or updated incident becomes visible without requiring the operator to sign in again.

### 1.3 Maintain field units

**Used by:** system administrator for setup; field unit or operator during a shift.

**Inputs:**

- Stable unit call sign.
- Unit type and capabilities.
- Home zone.
- Current shift members, optional in Stage 1.
- Availability: off duty, available, temporarily unavailable, or out of service.
- Optional unavailability reason and expected return time.

**Outputs:**

- Current unit record.
- Time and author of the latest availability change.
- Validation errors for unknown capabilities or illegal changes.

**Rules:**

- Call signs are unique among active units.
- A unit with an active assignment cannot become off duty without supervisor intervention.
- Removing a capability does not rewrite completed incident history.

### 1.4 Manually assign a unit

**Used by:** emergency operator.

**Inputs:**

- Incident identifier.
- Unit identifier.
- Assignment role: primary, support, medical support, perimeter, transport, or other.
- Operator's reason, optional for routine assignment and required when assigning an unavailable or unsuitable unit.
- Expected incident version shown to the operator.
- Client-generated command identifier.

**Outputs:**

- Confirmed assignment with assignment time and status.
- Updated incident status.
- Updated unit availability.
- Timeline event describing the decision.
- On conflict, the current incident and unit state plus a reason the command was rejected.

**Rules:**

- A unit cannot hold two incompatible active assignments.
- An operator cannot normally assign an off-duty or out-of-service unit.
- A repeated command identifier returns the original assignment result.
- If another operator changed the incident after it was displayed, the assignment is rejected as stale rather than overwriting the newer decision.

### 1.5 Progress an assignment

**Used by:** assigned field unit; operator as an explicitly attributed fallback.

**Inputs:**

- Assignment identifier.
- New status: accepted, en route, arrived, work started, completed, unable to complete, or cancelled.
- Optional note.
- Client-generated command identifier.
- Occurrence time when reporting an action after temporary loss of connectivity.

**Outputs:**

- Accepted status and recorded time.
- Current assignment and incident status.
- Timeline entry.
- If rejected, the allowed next actions and current state.

**Rules:**

- Status transitions follow an explicit lifecycle; a unit cannot arrive before accepting or being assigned.
- Repeating an already accepted transition returns success without adding duplicate events.
- A late report records both occurrence time and receipt time.
- Completing the primary assignment does not automatically close the incident when other required work remains.

### 1.6 Close an incident

**Used by:** emergency operator or shift supervisor.

**Inputs:**

- Incident identifier.
- Outcome: resolved, false alarm, duplicate, referred to another agency, unable to resolve, or cancelled.
- Closing summary.
- Optional linked incident when the outcome is duplicate.

**Outputs:**

- Closed incident record.
- Final timeline.
- Released units and their resulting availability.
- Resident-safe final status text.

**Rules:**

- An incident with active assignments requires confirmation before closure.
- Duplicate incidents remain independently auditable and point to the retained incident.
- Reopening requires a reason and creates a new timeline event.

### Stage 1 acceptance scenarios

1. An operator creates a fire report, assigns Engine 12, follows it through arrival and completion, and closes the incident.
2. Retrying report creation with the same submission identifier returns one incident.
3. Two operators try to assign the same unit to incompatible incidents; only one assignment succeeds and the other receives current state.
4. An operator cannot skip an illegal assignment transition.
5. An auditor reconstructs the complete sequence without consulting mutable application logs.

---

## Stage 2 — Offers, automated recommendations, and supervision

### Goal

Reduce operator search time while keeping humans in control. The product recommends suitable units and manages time-limited offers, rejection, reassignment, and supervisor overrides.

### 2.1 Report unit position and readiness

**Used by:** field-unit device or simulator.

**Inputs:**

- Unit identifier.
- Latitude and longitude.
- Device-recorded time.
- Increasing sequence number for the unit's current device session.
- Accuracy estimate.
- Movement heading and speed, optional.
- Readiness: available, occupied, temporarily unavailable, or emergency assistance requested.

**Outputs:**

- Accepted, duplicate, stale, or invalid result.
- Last accepted sequence number and time.
- Current server view of unit readiness.

**Rules:**

- A position older than the latest accepted update cannot replace the current position.
- An exact retry is acknowledged without creating another movement record.
- Impossible coordinates and implausible jumps are rejected or marked for review.
- A unit whose last position is too old is shown as location unknown, not nearby.

### 2.2 Request response recommendations

**Used by:** emergency operator.

**Inputs:**

- Incident identifier.
- Required capabilities and number of units by role.
- Maximum acceptable response distance or time, optional.
- Units or zones to exclude, with reasons.

**Outputs:**

- Ranked candidates.
- For each candidate: call sign, capabilities, readiness, location freshness, estimated distance/time, current workload, and human-readable inclusion or exclusion reasons.
- Warning when requirements cannot be satisfied.
- Recommendation generation time and the incident version used.

**Rules:**

- The product never presents stale or ineligible units as fully suitable without a warning.
- Ranking is deterministic for identical inputs and operational state.
- The operator can inspect why a higher-ranked unit was preferred.
- A recommendation does not reserve or assign a unit.

### 2.3 Send assignment offers

**Used by:** emergency operator, using selected recommendations.

**Inputs:**

- Incident identifier and role to fill.
- One or more selected units.
- Offer expiration duration.
- Offer strategy: sequential or simultaneous.
- Optional operator message.

**Outputs:**

- Offer identifiers and expiration times.
- Per-unit delivery state: pending delivery, delivered, unavailable, or failed.
- Incident board update showing that a response is being sought.

**Rules:**

- A sequential strategy offers to the next candidate only after rejection or expiration.
- A simultaneous strategy may contact several units, but only the required number can win assignments.
- Expired offers cannot be accepted.
- Sending the same request again does not create additional live offers.

### 2.4 Accept or reject an offer

**Used by:** field unit.

**Inputs:**

- Offer identifier.
- Decision: accept or reject.
- Rejection reason: occupied, unsafe, equipment problem, too far, ending shift, or other.
- Optional note.
- Client-generated command identifier.

**Outputs:**

- If accepted: confirmed assignment and incident summary needed by the unit.
- If rejected: acknowledgement and restored unit readiness where appropriate.
- If no longer valid: exact reason such as expired, withdrawn, already filled, or unit no longer eligible.

**Rules:**

- Concurrent acceptances cannot create more assignments than requested.
- An acceptance received after cancellation or expiry does not create an assignment.
- Rejection reasons become available for operational analysis.

### 2.5 Supervisor override

**Used by:** shift supervisor.

**Inputs:**

- Target incident, assignment, offer, priority, or unit state.
- Requested override.
- Mandatory reason.
- Optional expiry time for temporary overrides.

**Outputs:**

- Resulting state.
- Explicit list of warnings bypassed.
- Timeline event attributed to the supervisor.
- Notification to affected operator and units.

**Rules:**

- An override cannot erase earlier decisions.
- Some configured actions may require a second supervisor's approval.
- Temporary overrides revert or request review at expiry; they do not remain silently active.

### 2.6 Request Batman's assistance

**Used by:** shift supervisor or another explicitly authorized command role.

**Inputs:**

- Incident identifier.
- Assistance requested: reconnaissance, rescue, pursuit support, hazardous access, or other exceptional capability.
- Urgency and safe rendezvous/contact instructions.
- Minimum necessary incident summary.
- Reason ordinary city resources are insufficient.
- Expiration time for the request.

**Outputs:**

- Request reference and status: prepared, transmitted, acknowledged, declined, expired, or assistance in progress.
- Optional estimated arrival window supplied by Batman.
- Operational updates deliberately shared by Batman.
- Timeline record visible only to authorized roles.

**Rules:**

- Creating a request does not mean Batman accepted it or was assigned.
- The transmitted payload contains only information necessary for the requested assistance.
- The system must not require a legal name, permanent location, or ordinary employee/unit record for Batman.
- Batman may acknowledge, decline, or end assistance; supervisors retain command of official city units.
- Operators receive a safe public status such as “specialized assistance requested,” never Batman's private operational details.
- A request expires visibly if no acknowledgement arrives; it must not remain indefinitely pending.

### Stage 2 acceptance scenarios

1. The product ranks three medical units and explains why the nearest unit is excluded because its location is stale.
2. Two units accept simultaneous offers for one primary role; one receives the assignment and the other receives “role already filled.”
3. A rejected offer advances a sequential offer to the next unit.
4. A supervisor assigns an otherwise unsuitable unit, and the warning plus justification remains visible in the timeline.
5. Out-of-order location reports do not move a unit back to an old position.
6. A supervisor activates the authorized Batman channel, receives an acknowledgement without learning a secret identity or permanent location, and can distinguish requested assistance from confirmed assistance.

---

## Stage 3 — Live operations and resilient field communication

### Goal

Allow operators, supervisors, and field units to follow active incidents without repeatedly refreshing, while handling disconnections, slow clients, missed updates, and confidential information correctly.

### 3.1 Subscribe to live operational updates

**Used by:** operator console, supervisor console, and field-unit client.

**Inputs:**

- Subscription scope: assigned incidents, owned incidents, selected zones, selected incident, or city-wide summary.
- Last received event position, optional when reconnecting.
- Client connection identifier.

**Outputs:**

- Initial authorized snapshot.
- Ordered updates containing event identifier, position, occurrence time, entity reference, event type, and permitted payload.
- Heartbeat/liveness information.
- Explicit instruction to reload a snapshot when missed history is no longer available.

**Rules:**

- A reconnecting client can request updates after its last confirmed position.
- A client never receives details it is not currently authorized to see.
- Losing permission stops future confidential updates and invalidates affected subscriptions.
- A slow client is warned and eventually disconnected instead of delaying all other users.
- Duplicate updates may be delivered, but each carries a stable identifier so the client can ignore repeats.

### 3.2 Maintain an incident operations room

**Used by:** assigned units, owning operator, and supervisors.

**Inputs:**

- Incident identifier.
- Message type: operational note, request for information, safety warning, arrival update, or command acknowledgement.
- Text and optional attachment reference.
- Optional reply-to message.
- Client-generated message identifier.

**Outputs:**

- Accepted message with authoritative time and author.
- Delivery/read acknowledgements where supported.
- Ordered room history for authorized participants.

**Rules:**

- Joining an incident does not automatically grant access to sensitive historical content.
- Editing is limited to a short correction window and preserves the original version.
- Deletion removes content from ordinary display only when policy permits; audit metadata remains.
- Safety warnings are visually distinct and require acknowledgement from targeted units.

### 3.3 Operate through temporary disconnection

**Used by:** field-unit client.

**Inputs:**

- A batch of locally recorded actions with stable identifiers, local order, and occurrence times.
- The client's last known assignment version.

**Outputs:**

- Per-action result: applied, already applied, rejected due to conflict, invalid, or requires human review.
- Current authoritative assignment state.
- Missing updates since the client's last known position.

**Rules:**

- Reconnecting does not blindly replay actions over newer supervisor decisions.
- Independent valid actions may succeed even if another action in the batch conflicts.
- Conflicts remain visible to the operator and field unit until acknowledged.

### Stage 3 acceptance scenarios

1. An operator disconnects, reconnects with the last received position, and obtains every missed incident update in order.
2. A unit records arrival while offline; after reconnection the action is accepted with distinct occurrence and receipt times.
3. An offline “accept offer” action conflicts with a supervisor cancellation and is rejected without resurrecting the assignment.
4. A city-wide supervisor view continues updating while a deliberately slow client is disconnected.
5. Revoking a user's zone access stops confidential updates on an existing connection.

---

## Stage 4 — Gotham under pressure

### Goal

Support coordinated response during a major event when reports surge, information is incomplete, incidents are related, and normal capacity is insufficient.

### 4.1 Declare and manage an emergency operation

**Used by:** shift supervisor.

**Inputs:**

- Operation name and summary.
- Affected zones and expected duration.
- Severity level.
- Command roles and participating agencies.
- Temporary policies, such as priority rules or reserved units.

**Outputs:**

- Emergency-operation identifier and status.
- Shared operational view containing linked incidents, resource demand, and key decisions.
- Notifications to affected users.

**Rules:**

- Declaring an operation does not silently change existing incident priorities.
- Temporary policy changes show their author, reason, scope, and expiry.
- Ending an operation produces a list of unresolved incidents and active temporary policies requiring disposition.

### 4.2 Link related reports and incidents

**Used by:** emergency operator and supervisor.

**Inputs:**

- Two or more reports/incidents.
- Relationship: duplicate, same event, consequence of, nearby but independent, or uncertain.
- Reason and confidence.

**Outputs:**

- Relationship visible on relevant incident views.
- Combined summary for a declared emergency operation.
- Warning about conflicting locations, categories, or reported facts.

**Rules:**

- Linking does not merge timelines or erase identifiers.
- Unlinking records who changed the assessment and why.
- Residents continue tracking the public reference they were originally given.

### 4.3 Triage reports during a surge

**Used by:** emergency operator.

**Inputs:**

- Incoming reports from residents, sensors, and partner agencies.
- Answers to configurable triage questions.
- Optional connection to an emergency operation.

**Outputs:**

- Suggested priority, category, and possible related incidents.
- Queue placement and estimated review urgency.
- Explicit reasons for the suggestion.
- Warning when required information is missing or contradictory.

**Rules:**

- A suggestion never closes or discards a report without human action.
- Reports involving immediate danger or trapped people cannot be deprioritized solely because the queue is large.
- The operator's departure from a suggestion is recorded but does not require punitive justification for routine judgment.

### 4.4 Reserve and rebalance resources

**Used by:** shift supervisor.

**Inputs:**

- Units or capability capacity to reserve.
- Reserved zones or emergency operation.
- Start, expiry, and reason.
- Rebalancing instruction moving available coverage between zones.

**Outputs:**

- Updated resource plan.
- Incidents and zones whose expected coverage worsens.
- Recommendations affected by the reservation.
- Notifications to responsible operators.

**Rules:**

- Reserved units are excluded from routine recommendations but remain visible with the reason.
- Existing assignments are not cancelled automatically.
- A reservation cannot remain indefinitely without renewal.

### 4.5 Degraded-mode operation

**Used by:** all operational roles.

**Inputs:**

- Normal user actions while one or more product capabilities are unavailable or delayed.

**Outputs:**

- Clear indication of which data may be stale.
- A safe subset of actions that remains available.
- Confirmation that an action was accepted, queued for later processing, or not accepted.
- Recovery summary after normal operation returns.

**Rules:**

- The interface must never claim an assignment is confirmed when confirmation is uncertain.
- Critical commands receive a stable receipt that can later be reconciled.
- Recovery does not create duplicate incidents or assignments.
- Operators can inspect unresolved conflicts caused by the degraded period.

### Stage 4 acceptance scenarios

1. A bridge collapse generates many related reports; operators link them without destroying their independent histories.
2. A supervisor reserves medical capacity for the affected zone, and routine recommendations explain why those units are unavailable.
3. During a dependency outage, an operator can see that candidate locations are stale and chooses a manual unit knowingly.
4. Actions submitted during degraded operation are reconciled without duplicate assignments.
5. Ending the emergency operation exposes unresolved incidents and temporary policies instead of hiding them.

---

## Stage 5 — Accountability, resident communication, and learning

### Goal

Make the system useful after the immediate response: residents receive appropriate updates, supervisors review outcomes, auditors reconstruct decisions, and Gotham improves future operations using measured evidence.

### 5.1 Resident status tracking

**Used by:** resident with a public report reference and private access code.

**Inputs:**

- Public report reference.
- Private access code or verified contact method.

**Outputs:**

- Resident-safe status: received, under review, response dispatched, responders arrived, resolved, or more information required.
- Last update time.
- Safe request for additional information, when applicable.
- Closure summary appropriate for public disclosure.

**Rules:**

- Exact responder location, identities, internal notes, other reporters, and tactical details are never exposed.
- A linked or duplicate report retains a meaningful resident-facing status.
- Sensitive incidents may expose only acknowledgement and a contact channel.

### 5.2 Request additional information

**Used by:** emergency operator and resident.

**Operator inputs:**

- Report reference.
- Question or structured information request.
- Response deadline and urgency.

**Resident inputs:**

- Text response.
- Optional attachment.
- Confirmation that the incident is still occurring.

**Outputs:**

- Recorded request and response.
- Operator notification.
- Timeline entry showing when resident information influenced the incident.

**Rules:**

- Resident content is treated as unverified until reviewed.
- Attachments are not automatically trusted or publicly visible.
- Failure to respond does not automatically close a high-priority incident.

### 5.3 Conduct an incident review

**Used by:** shift supervisor and invited participants.

**Inputs:**

- Incident or emergency operation.
- Review scope and participants.
- What went well, what was confusing, delays, safety concerns, and proposed follow-up actions.
- References to relevant timeline events.

**Outputs:**

- Review document linked to the incident.
- Measured response intervals and important decision points.
- Action items with owner, due date, and status.
- Explicit corrections to earlier assumptions without rewriting history.

**Rules:**

- Measured intervals distinguish event occurrence, system receipt, operator action, and unit acknowledgement.
- Review access can be narrower than ordinary incident access.
- Closing the review does not automatically close its action items.

### 5.4 Search the audit history

**Used by:** auditor and authorized supervisor.

**Inputs:**

- Time range.
- Actor, incident, unit, operation, action type, or public reference.
- Optional before/after state criteria.
- Stated audit purpose.

**Outputs:**

- Ordered matching events.
- Actor, time, action, reason, and permitted before/after values.
- Evidence of overrides, failed attempts, corrections, and access-sensitive redactions.
- Export with generation time and requesting auditor.

**Rules:**

- Audit searches and exports are themselves audited.
- Redacted values remain visibly marked as redacted rather than appearing absent.
- Ordinary administrators cannot modify audit results.

### 5.5 Operational performance report

**Used by:** supervisor and operations analyst.

**Inputs:**

- Time period, zones, categories, priorities, and emergency operations.
- Comparison period, optional.

**Outputs:**

- Time from report to operator review.
- Time from validated incident to first offer.
- Offer acceptance and expiration rates.
- Time from assignment to en route and arrival.
- Reassignment frequency and reasons.
- Percentage of incidents with stale or missing location data.
- Outcome distribution.
- Data-quality warnings and excluded samples.

**Rules:**

- Percentiles are shown in addition to averages where timing is reported.
- Reports state which population and time boundaries were used.
- Missing events are disclosed rather than interpreted as zero duration.
- Metrics support operational improvement and are not presented as proof of individual employee quality without context.

### Stage 5 acceptance scenarios

1. A resident follows a report without learning the responder's identity or exact location.
2. A supervisor reviews a delayed response and distinguishes a late field update from a genuinely late arrival.
3. An auditor finds the supervisor override, the warning it bypassed, and the later correction.
4. An audit export records who generated it and why.
5. A monthly report shows response-time percentiles, reassignment reasons, and missing-data warnings.

---

## Stage 6 — Multi-agency Gotham

### Goal

Coordinate independent Gotham agencies without pretending they share identical permissions, terminology, or operational control.

### 6.1 Partner-agency incident exchange

**Used by:** emergency operator and approved partner system.

**Inputs:**

- Partner agency identity.
- Partner's stable report identifier.
- Category, priority, location, description, and contact channel.
- Information-sharing classification.
- Stable submission identifier.

**Outputs:**

- Accepted, rejected, duplicate, or requires-human-review result.
- Gotham incident reference when accepted.
- Mapping between partner and Gotham identifiers.
- Field-level warnings for information that could not be interpreted.

**Rules:**

- Repeated partner submissions do not create duplicate incidents.
- Unknown categories are preserved and routed for human mapping rather than silently discarded.
- Partner data is visible only according to its sharing classification.

### 6.2 Request and commit external assistance

**Used by:** Gotham supervisor and partner-agency coordinator.

**Inputs:**

- Incident or emergency operation.
- Requested capability, quantity, urgency, staging location, and expected duration.
- Constraints on information sharing.

**Outputs:**

- Assistance request with pending, accepted, partially accepted, rejected, cancelled, or completed status.
- Committed external resources at the level the partner permits Gotham to see.
- Timeline of decisions by both organizations.

**Rules:**

- A requested external resource is not shown as assigned until the partner commits it.
- Each organization retains control of its own resources.
- Cancellation and modification races produce explicit outcomes and remain auditable.

### 6.3 Shared situational summary

**Used by:** authorized users across participating agencies.

**Inputs:**

- Emergency operation.
- User's agency and permissions.
- Optional zone, incident category, or time filter.

**Outputs:**

- Common overview of incidents, needs, committed capabilities, major hazards, and decisions.
- Agency-specific redactions.
- Data freshness for each contributing agency.

**Rules:**

- The summary never implies that stale partner data is current.
- Users may see different permitted details while sharing the same stable incident references.
- Loss of a partner connection leaves the last known contribution visibly marked as stale.

### Stage 6 acceptance scenarios

1. A hospital submits the same emergency report three times and receives one Gotham incident mapping.
2. Gotham requests two medical teams; the partner commits one, and the product shows partial acceptance rather than success.
3. A partner revokes access to tactical details, and existing users stop receiving those details without losing the shared incident reference.
4. A disconnected agency's information remains visible but clearly marked with its last update time.

---

## Cross-stage product evidence

For each completed stage, preserve a sanitized product packet containing:

- The user problem and scope.
- Functional decisions and rejected alternatives.
- Example inputs and outputs.
- Acceptance scenarios and test evidence.
- Feedback from a human reviewer.
- A requirement changed after feedback or observed behavior.
- A failure or ambiguity discovered during implementation.
- Measured user-facing or operational result.

The capstone is successful when it tells a coherent story of increasing ownership: first a reliable incident workflow, then concurrent dispatch, live operation, crisis behavior, accountability, and finally coordination across organizational boundaries.
