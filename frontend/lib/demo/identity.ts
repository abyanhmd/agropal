import {DEMO_FARM_ID} from "@/lib/demo/base";

// Who the demo pretends to be. Small, so the signed-in app can import it without the sample data.

export const DEMO_USER = {
    user_id: "demo-user",
    name: "Maya Whitfield",
    email: "maya@maplehollow.example",
    username: "maya",
    avatar: "",
    role: "USER",
};

export const DEMO_FARM = {farm_id: DEMO_FARM_ID, user_id: DEMO_USER.user_id, name: "Maple Hollow Farm"};

// Central Iowa, inside the crop calendar's coverage
export const DEMO_LOCATION = {location_id: "demo-location", latitude: 42.03, longitude: -93.62};
export const DEMO_LOCATION_NAME = "Ames, Iowa";
