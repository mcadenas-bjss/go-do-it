import type { APIRoute } from "astro";
import { BASE_URL, TOGGLE_TODO_PATH, TODO_PATH } from "../../../../utils/globals";

export const POST: APIRoute = async ({params}) => {
    try {
        const id = params.id;
        const res = await fetch(`${BASE_URL}${TOGGLE_TODO_PATH}/${id}`, {method: "POST"});
        if (!res.ok) throw new Error("Error");
        return new Response(null, { status: 200 })
    } catch (e) {
        return new Response(
            JSON.stringify({
                message: "An error occurred.",
            }),
            {
                status: 500,
            }
        );
    }
}