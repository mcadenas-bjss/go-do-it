import type { APIRoute } from "astro";
import { BASE_URL, TODO_PATH } from "../../../utils/globals";

export const POST: APIRoute = async ({ params, request }) => {
    console.log('POST /api/todo');
    try {
        const form = await request.formData()
        const Description = form.get("description") as string
        const Time = form.get("time") as string
        const body: Todo = {
            Id: 0,
            Time: Time ? new Date(Time).toISOString(): "",
            Description,
            Completed: false,
        }
        const todo = await fetch(`${BASE_URL}${TODO_PATH}`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(body),
        }).then(res => {
            return res.body});

        return new Response(todo, {status: 200})
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