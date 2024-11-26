import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';

@Component({
  selector: 'app-magic-link',
  templateUrl: './magic-link.component.html',
  styleUrls: ['./magic-link.component.css'],
})
export class MagicLinkComponent implements OnInit {
  token: string | null = null;
  loading: boolean = false;
  errorMessage: string = '';

  constructor(private route: ActivatedRoute, private router: Router) {}

  ngOnInit(): void {
    this.route.queryParams.subscribe((params) => {
      this.token = params['token'];
      if (this.token) {
        this.storeTokenAndNavigate();
      } else {
        this.errorMessage = 'Invalid or missing token.';
      }
    });
  }

  storeTokenAndNavigate() {
    if (this.token) {
      localStorage.setItem('jwt', this.token);
      this.router.navigate(['/all-projects']);
    } else {
      this.errorMessage = 'Invalid token';
    }
  }
}
