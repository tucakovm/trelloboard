import { Injectable } from '@angular/core';
import { CanActivate, Router, ActivatedRouteSnapshot } from '@angular/router';
import { AuthService } from './services/auth.service';

@Injectable({
  providedIn: 'root',
})
export class AuthGuard implements CanActivate {
  constructor(private router: Router, private authService:AuthService){}

  canActivate(route: ActivatedRouteSnapshot): boolean {
    const isAuthenticated = this.authService.tokenIsPresent();
    const userRoles = this.authService.getUserRoles();
    const requiredRoles = route.data['roles'] as string[];
    console.log(userRoles)

    if (!isAuthenticated) {
      this.router.navigate(['/login']); 
      return false;
    }

    if (requiredRoles && !this.hasRequiredRole(userRoles, requiredRoles)) {
      this.router.navigate(['/unauthorized']); 
      return false;
    }

    return true;
  }

  private hasRequiredRole(userRoles: string[], requiredRoles: string[]): boolean {
    // Provera da li korisnik ima barem jednu od potrebnih uloga
    return requiredRoles.some((role) => userRoles.includes(role));
  }
}