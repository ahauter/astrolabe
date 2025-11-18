"""
Subtract b from a.

Parameters:
a  (float): The number to subtract from.
b  (float): The number to subtract.

Returns:
float: a minus b.
"""


def add(a, b):
    """
    Add two numbers.

    Parameters:
        a(float): The first number.
        b(float): The second number.

    Returns:
        float: The sum of a and b.
    """
    return a + b


def subtract(a, b):
    """
    Subtract one number from another.

    Parameters:
        a(float): The number to subtract from .
        b(float): The number to subtract.

    Returns:
        float: The result of a - b.
    """
    # a second comment
    return a - b


def multiply(a, b):
    """
    Multiply two numbers.

    Parameters:
        a(float): The first number.
        b(float): The second number.

    Returns:
        float: The product of a and b.
    """
    return a * b  # a second comment


def divide(a, b):
    """
    Divide one number by another.

    Parameters:
        a(float): The numerator.
        b(float): The denominator.

    Returns:
        float: The result of a / b.

    Raises:
        ValueError: If b is zero.
    """
    if b == 0:
        raise ValueError("Cannot divide by zero.")
    return a / b


def factorial(n):
    """
    Compute the factorial of a non-negative integer.

    Parameters:
        n(int): The number to compute the factorial for .

    Returns:
        int: The factorial of n.

    Raises:
        ValueError: If n is negative.
    """
    if n < 0:
        raise ValueError("Factorial is not defined for negative numbers.")
    if n == 0:
        return 1
    return n * factorial(n - 1)
