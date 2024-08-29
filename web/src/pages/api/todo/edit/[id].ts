import type { APIRoute } from "astro";

export const GET: APIRoute = async ({params}) => {
    try {
        const id = params.id
        const html = `<li id="todo-${id}"}>
            <form class="todo" action="submit" method="update">
              <div>
                <label for="description">Description:</label>
                <textarea
                  id="description"
                  name="description"
                  rows="1"
                  cols="50"
                  required
                /></textarea>
              </div>
              <div>
                <label for="time">Time:</label>
                <input
                  type="datetime-local"
                  id="time"
                  name="time"
                  min="2018-06-07T00:00"
                />
              </div>
              <button
                id="save-${id}"
                hx-post="/api/todo/${id}"
                hx-swap="delete"
                hx-target="this"
              >
                Save
              </button>
            </form>
          </li>`
        return new Response(html, { status: 200, headers: {
            "Content-Type": "text/html"
        } })
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