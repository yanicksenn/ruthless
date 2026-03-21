import { GrpcWebFetchTransport } from "@protobuf-ts/grpcweb-transport";
import { CardServiceClient, DeckServiceClient, SessionServiceClient, GameServiceClient, UserServiceClient, FriendServiceClient, NotificationServiceClient, SessionInvitationServiceClient } from "./ruthless.client";
import type { RpcOptions } from "@protobuf-ts/runtime-rpc";

const transport = new GrpcWebFetchTransport({
  baseUrl: import.meta.env.VITE_API_BASE_URL || "http://localhost:8080",
});

// Helper to add auth header
const createOptions = (token: string | null): RpcOptions => {
  const meta: Record<string, string> = {};
  if (token) {
    meta["Authorization"] = `Bearer ${token}`;
  }
  return { meta };
};

export const cardClient = new CardServiceClient(transport);
export const deckClient = new DeckServiceClient(transport);
export const sessionClient = new SessionServiceClient(transport);
export const gameClient = new GameServiceClient(transport);
export const userClient = new UserServiceClient(transport);
export const friendClient = new FriendServiceClient(transport);
export const notificationClient = new NotificationServiceClient(transport);
export const sessionInvitationClient = new SessionInvitationServiceClient(transport);

export { createOptions };
