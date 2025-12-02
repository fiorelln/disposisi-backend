import 'dart:convert';
import 'package:http/http.dart' as http;

class ApiService {
  static const _base = 'http://10.0.2.2:7000/auth';

  static Future<Map<String, dynamic>> register({
    required String name,
    required String email,
    required String password,
    required String role,
  }) async {
    final uri = Uri.parse('$_base/register');
    final res = await http.post(
      uri,
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({
        'name': name,
        'email': email,
        'password': password,
        'role': role,
      }),
    );

    final body = res.body.isNotEmpty ? jsonDecode(res.body) : {};
    return {
      'success': res.statusCode == 200,
      'statusCode': res.statusCode,
      'data': body,
    };
  }
  
  // validate email and password
  static Future<Map<String, dynamic>> login({
    required String email,
    required String password,
  }) async {
    final uri = Uri.parse('$_base/login');
    final res = await http.post(
      uri,
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({'email': email, 'password': password}),
    );

    final body = res.body.isNotEmpty ? jsonDecode(res.body) : {};
    return {
      'success': res.statusCode == 200,
      'statusCode': res.statusCode,
      'data': body,
    };
  }
}
