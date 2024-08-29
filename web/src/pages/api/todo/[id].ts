import type { APIRoute } from "astro";
import { BASE_URL, TODO_PATH, TODOS_PATH } from "../../../utils/globals";

export const GET: APIRoute = async ({params}) => {
    try {
        const id = params.id
        const response = await fetch(`${BASE_URL}${TODOS_PATH}/${id}`);
        const data: Todos = await response.json();
        return new Response(JSON.stringify(data), { status: 200 })
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

export const POST: APIRoute = async ({ params, request }) => {
    try {
        const id = params.id
        if (!id) throw new Error("Id is required")
        const form = await request.formData()
        const Description = form.get("description") as string
        const Time = form.get("time") as string
        const body: Todo = {
            Id: parseInt(id, 10),
            Time: Time ? new Date(Time).toISOString(): "",
            Description,
            Completed: false,
        }
        await fetch(`${BASE_URL}${TODO_PATH}/${id}`, {
            body: JSON.stringify(body),
            method: "PUT",
            headers: {
                "Content-Type": "application/json",
            },
        })

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

export const DELETE: APIRoute = async ({ params }) => {
    try {
        const id = params.id
        await fetch(`${BASE_URL}${TODO_PATH}/${id}`, {
            method: "DELETE",
        })

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