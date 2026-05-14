import { useEffect, useState } from "react";
import { api } from "./api";
import "./style.css";

export default function App() {
  const [users, setUsers] = useState([]);
  const [books, setBooks] = useState([]);
  const [borrows, setBorrows] = useState([]);
  const [message, setMessage] = useState("");

  const [userForm, setUserForm] = useState({
    name: "",
    email: "",
    password: "",
  });

  const [bookForm, setBookForm] = useState({
    name: "",
    authors: "",
    year: "",
  });

  const [borrowForm, setBorrowForm] = useState({
    user_id: "",
    book_id: "",
  });

  async function loadData() {
    try {
      const [usersData, booksData, borrowsData] = await Promise.allSettled([
        api.listUsers(),
        api.listBooks(),
        api.listBorrows(),
      ]);

      if (usersData.status === "fulfilled") {
        setUsers(usersData.value.users || usersData.value || []);
      }

      if (booksData.status === "fulfilled") {
        setBooks(booksData.value.books || booksData.value || []);
      }

      if (borrowsData.status === "fulfilled") {
        setBorrows(borrowsData.value.borrows || borrowsData.value || []);
      }
    } catch (error) {
      setMessage(error.message);
    }
  }

  useEffect(() => {
    loadData();
  }, []);

  async function handleCreateUser(e) {
    e.preventDefault();

    try {
      await api.createUser(userForm);
      setUserForm({ name: "", email: "", password: "" });
      setMessage("User created successfully");
      await loadData();
    } catch (error) {
      setMessage(error.message);
    }
  }

  async function handleCreateBook(e) {
    e.preventDefault();

    try {
      await api.createBook({
        name: bookForm.name,
        authors: bookForm.authors,
        year: Number(bookForm.year),
      });

      setBookForm({ name: "", authors: "", year: "" });
      setMessage("Book created successfully");
      await loadData();
    } catch (error) {
      setMessage(error.message);
    }
  }

  async function handleBorrowBook(e) {
    e.preventDefault();

    try {
      await api.borrowBook(borrowForm);
      setBorrowForm({ user_id: "", book_id: "" });
      setMessage("Book borrowed successfully");
      await loadData();
    } catch (error) {
      setMessage(error.message);
    }
  }

  async function handleReturnBook(borrowId) {
    try {
      await api.returnBook(borrowId);
      setMessage("Book returned successfully");
      await loadData();
    } catch (error) {
      setMessage(error.message);
    }
  }

  return (
    <div className="app">
      <header className="header">
        <div>
          <h1>E-Library Management System</h1>
          <p>Microservices frontend for User, Book and Borrow services</p>
        </div>
        <button onClick={loadData}>Refresh</button>
      </header>

      {message && <div className="message">{message}</div>}

      <main className="grid">
        <section className="card">
          <h2>Create User</h2>
          <form onSubmit={handleCreateUser}>
            <input
              placeholder="Name"
              value={userForm.name}
              onChange={(e) =>
                setUserForm({ ...userForm, name: e.target.value })
              }
              required
            />
            <input
              placeholder="Email"
              type="email"
              value={userForm.email}
              onChange={(e) =>
                setUserForm({ ...userForm, email: e.target.value })
              }
              required
            />
            <input
              placeholder="Password"
              type="password"
              value={userForm.password}
              onChange={(e) =>
                setUserForm({ ...userForm, password: e.target.value })
              }
              required
            />
            <button type="submit">Create User</button>
          </form>
        </section>

        <section className="card">
          <h2>Create Book</h2>
          <form onSubmit={handleCreateBook}>
            <input
              placeholder="Book name"
              value={bookForm.name}
              onChange={(e) =>
                setBookForm({ ...bookForm, name: e.target.value })
              }
              required
            />
            <input
              placeholder="Authors"
              value={bookForm.authors}
              onChange={(e) =>
                setBookForm({ ...bookForm, authors: e.target.value })
              }
              required
            />
            <input
              placeholder="Year"
              type="number"
              value={bookForm.year}
              onChange={(e) =>
                setBookForm({ ...bookForm, year: e.target.value })
              }
              required
            />
            <button type="submit">Create Book</button>
          </form>
        </section>

        <section className="card">
          <h2>Borrow Book</h2>
          <form onSubmit={handleBorrowBook}>
            <input
              placeholder="User ID"
              value={borrowForm.user_id}
              onChange={(e) =>
                setBorrowForm({ ...borrowForm, user_id: e.target.value })
              }
              required
            />
            <input
              placeholder="Book ID"
              value={borrowForm.book_id}
              onChange={(e) =>
                setBorrowForm({ ...borrowForm, book_id: e.target.value })
              }
              required
            />
            <button type="submit">Borrow</button>
          </form>
        </section>
      </main>

      <section className="tables">
        <DataTable title="Users" items={users} />
        <DataTable title="Books" items={books} />
        <BorrowTable items={borrows} onReturn={handleReturnBook} />
      </section>
    </div>
  );
}

function DataTable({ title, items }) {
  return (
    <section className="card wide">
      <h2>{title}</h2>

      {items.length === 0 ? (
        <p className="empty">No data</p>
      ) : (
        <div className="list">
          {items.map((item, index) => (
            <pre key={item.id || item.user_id || item.book_id || index}>
              {JSON.stringify(item, null, 2)}
            </pre>
          ))}
        </div>
      )}
    </section>
  );
}

function BorrowTable({ items, onReturn }) {
  return (
    <section className="card wide">
      <h2>Borrows</h2>

      {items.length === 0 ? (
        <p className="empty">No borrows</p>
      ) : (
        <div className="list">
          {items.map((borrow, index) => {
            const borrowId = borrow.borrow_id || borrow.id;

            return (
              <div className="borrow-row" key={borrowId || index}>
                <pre>{JSON.stringify(borrow, null, 2)}</pre>
                {borrowId && borrow.status === "ACTIVE" && (
                  <button onClick={() => onReturn(borrowId)}>Return</button>
                )}
              </div>
            );
          })}
        </div>
      )}
    </section>
  );
}