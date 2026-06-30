// freebusy-client — browser SDK for the FreeBusy gRPC-Web API.
//
// Built on Connect-ES (@connectrpc/connect) + protobuf-es (@bufbuild/protobuf).
//
// Quick start:
//
//   import { createFreebusyClient } from "freebusy-client";
//
//   const fb = createFreebusyClient({ baseUrl: "https://api.freebusy.example" });
//
//   // Connect-ES methods take a plain message object and return the response
//   // message directly (no { response } wrapper):
//   const res = await fb.availability.computeAvailability({
//     resource: "resources/123",
//     // ...request fields
//   });
//
// Need a message/enum type? They are grouped per service:
//
//   import { booking } from "freebusy-client";
//   const req: booking.CreateBookingRequest = { /* ... */ };

import { createClient, type Client, type Transport } from "@connectrpc/connect";
import {
  createGrpcWebTransport,
  type GrpcWebTransportOptions,
} from "@connectrpc/connect-web";

import { AvailabilityService } from "./src/freebusy/availability/v1/availability_service_pb";
import { BookingService } from "./src/freebusy/booking/v1/booking_service_pb";
import { IdentityService } from "./src/freebusy/identity/v1/identity_service_pb";
import { OrganisationService } from "./src/freebusy/organisation/v1/organisation_service_pb";
import { PromoCodeService } from "./src/freebusy/promocode/v1/promocode_service_pb";
import { ResourceService } from "./src/freebusy/resource/v1/resource_service_pb";
import { ScheduleService } from "./src/freebusy/schedule/v1/schedule_service_pb";

// Client factory

/**
 * Options for {@link createFreebusyClient}. `baseUrl` is required; every other
 * field of the underlying {@link GrpcWebTransportOptions} (interceptors, custom
 * fetch, default headers, binary/text format, …) may be supplied to customise
 * the transport.
 */
export interface FreebusyClientOptions extends Partial<GrpcWebTransportOptions> {
  /** Base URL of the gRPC-Web endpoint, e.g. `https://api.freebusy.example`. */
  baseUrl: string;
}

/** A fully-wired set of FreeBusy service clients sharing one transport. */
export interface FreebusyClient {
  /** The shared transport. Reuse it to build additional clients if needed. */
  readonly transport: Transport;
  readonly availability: Client<typeof AvailabilityService>;
  readonly booking: Client<typeof BookingService>;
  readonly identity: Client<typeof IdentityService>;
  readonly organisation: Client<typeof OrganisationService>;
  readonly promoCode: Client<typeof PromoCodeService>;
  readonly resource: Client<typeof ResourceService>;
  readonly schedule: Client<typeof ScheduleService>;
}

/**
 * Create a FreeBusy client backed by a single gRPC-Web transport.
 *
 * Pass either {@link FreebusyClientOptions} (a transport is built for you) or a
 * pre-configured Connect {@link Transport} (advanced — e.g. to share one across
 * SDKs or to use a non-grpc-web protocol).
 */
export function createFreebusyClient(
  optionsOrTransport: FreebusyClientOptions | Transport,
): FreebusyClient {
  const transport: Transport =
    "baseUrl" in optionsOrTransport
      ? createGrpcWebTransport(optionsOrTransport)
      : optionsOrTransport;

  return {
    transport,
    availability: createClient(AvailabilityService, transport),
    booking: createClient(BookingService, transport),
    identity: createClient(IdentityService, transport),
    organisation: createClient(OrganisationService, transport),
    promoCode: createClient(PromoCodeService, transport),
    resource: createClient(ResourceService, transport),
    schedule: createClient(ScheduleService, transport),
  };
}

// Service descriptors + Connect primitives (for building clients yourself)

export { AvailabilityService } from "./src/freebusy/availability/v1/availability_service_pb";
export { BookingService } from "./src/freebusy/booking/v1/booking_service_pb";
export { IdentityService } from "./src/freebusy/identity/v1/identity_service_pb";
export { OrganisationService } from "./src/freebusy/organisation/v1/organisation_service_pb";
export { PromoCodeService } from "./src/freebusy/promocode/v1/promocode_service_pb";
export { ResourceService } from "./src/freebusy/resource/v1/resource_service_pb";
export { ScheduleService } from "./src/freebusy/schedule/v1/schedule_service_pb";

export { createClient, ConnectError, Code, type Client, type Transport, type Interceptor } from "@connectrpc/connect";
export { createGrpcWebTransport, type GrpcWebTransportOptions } from "@connectrpc/connect-web";

// Message & enum types, grouped per service (avoids cross-service name clashes)

export * as availability from "./types/availability";
export * as booking from "./types/booking";
export * as identity from "./types/identity";
export * as organisation from "./types/organisation";
export * as promocode from "./types/promocode";
export * as resource from "./types/resource";
export * as schedule from "./types/schedule";
export * as shared from "./types/shared";
